package role

import (
	"context"
	"fmt"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/internal/infrastructure/cache"
	"log"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (model.Role, error)
	FindAll(ctx context.Context) ([]model.Role, error)
	Create(ctx context.Context, m model.Role) (model.Role, error)
	Update(ctx context.Context, m model.Role) (model.Role, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo  Repository
	cache cache.Cache[any]
}

func NewService(repo Repository, cache cache.Cache[any]) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

const (
	roleCacheKeyPrefix = "role:"
	allRolesCacheKey   = "roles:all"
)

func (s *Service) GetByID(ctx context.Context, id string) (model.Role, error) {
	cacheKey := fmt.Sprintf("%s%s", roleCacheKeyPrefix, id)

	if val, found := s.cache.Get(ctx, cacheKey); found {
		if role, ok := val.(model.Role); ok {
			log.Printf("🚀 [Cache Hit] Role ID: %s", id)
			return role, nil
		}
	}

	log.Printf("📥 [Cache Miss] Fetching from DB - Role ID: %s", id)

	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Role{}, err
	}

	s.cache.Set(ctx, cacheKey, role, 10*time.Minute)

	return role, nil
}

func (s *Service) GetAll(ctx context.Context) ([]model.Role, error) {
	if val, found := s.cache.Get(ctx, allRolesCacheKey); found {
		if roles, ok := val.([]model.Role); ok {
			log.Printf("🚀 [Cache Hit] All Roles")
			return roles, nil
		}
	}

	log.Printf("📥 [Cache Miss] Fetching all roles from DB")

	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, allRolesCacheKey, roles, 5*time.Minute)

	return roles, nil
}

func (s *Service) Create(ctx context.Context, m model.Role) (model.Role, error) {
	role, err := s.repo.Create(ctx, m)
	if err != nil {
		return model.Role{}, err
	}

	s.cache.Delete(ctx, allRolesCacheKey)

	return role, nil
}

func (s *Service) Update(ctx context.Context, id string, m model.Role) (model.Role, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return model.Role{}, fmt.Errorf("invalid uuid: %w", err)
	}
	m.ID = parsedID

	role, err := s.repo.Update(ctx, m)
	if err != nil {
		return model.Role{}, err
	}

	s.cache.Delete(ctx, fmt.Sprintf("%s%s", roleCacheKeyPrefix, id))
	s.cache.Delete(ctx, allRolesCacheKey)

	return role, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	s.cache.Delete(ctx, fmt.Sprintf("%s%s", roleCacheKeyPrefix, id))
	s.cache.Delete(ctx, allRolesCacheKey)

	return nil
}
