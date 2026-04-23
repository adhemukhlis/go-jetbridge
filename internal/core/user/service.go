// internal/core/user/service.go
package user

import (
	"context"
	"fmt"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/internal/infrastructure/cache"
	"log"
	"time"

	"github.com/google/uuid"
)

// WithRoles represents a user with their associated roles.
type WithRoles struct {
	model.User
	Role []model.Role
}

// Repository defines the interface for data access operations.
type Repository interface {
	FindByID(ctx context.Context, id string) (WithRoles, error)
	FindAll(ctx context.Context) ([]WithRoles, error)
	Create(ctx context.Context, u model.User) (WithRoles, error)
	Update(ctx context.Context, u model.User) (WithRoles, error)
	Delete(ctx context.Context, id string) error
}

// Service provides high-level business logic for User operations.
type Service struct {
	repo  Repository
	cache cache.Cache[any]
}

// NewService creates a new instance of Service with the given repository and cache.
func NewService(repo Repository, cache cache.Cache[any]) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

const (
	userCacheKeyPrefix = "user:"
	allUsersCacheKey   = "users:all"
)

// GetByID retrieves a single user with their roles by their ID.
// It uses a cache-aside pattern to minimize database queries.
func (s *Service) GetByID(ctx context.Context, id string) (WithRoles, error) {
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, id)

	// Try to get from cache
	if val, found := s.cache.Get(ctx, cacheKey); found {
		if user, ok := val.(WithRoles); ok {
			log.Printf("🚀 [Cache Hit] User ID: %s", id)
			return user, nil
		}
	}

	log.Printf("📥 [Cache Miss] Fetching from DB - User ID: %s", id)

	// If not in cache, get from repo
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return WithRoles{}, err
	}

	// Save to cache for 10 minutes
	s.cache.Set(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// GetAll retrieves all users with their roles.
func (s *Service) GetAll(ctx context.Context) ([]WithRoles, error) {
	// Try to get from cache
	if val, found := s.cache.Get(ctx, allUsersCacheKey); found {
		if users, ok := val.([]WithRoles); ok {
			log.Printf("🚀 [Cache Hit] All Users")
			return users, nil
		}
	}

	log.Printf("📥 [Cache Miss] Fetching all users from DB")

	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache for 5 minutes
	s.cache.Set(ctx, allUsersCacheKey, users, 5*time.Minute)

	return users, nil
}

// Create creates a new user and invalidates the "all users" cache.
func (s *Service) Create(ctx context.Context, u model.User) (WithRoles, error) {
	user, err := s.repo.Create(ctx, u)
	if err != nil {
		return WithRoles{}, err
	}

	// Invalidate the list cache
	s.cache.Delete(ctx, allUsersCacheKey)

	return user, nil
}

// Update updates an existing user and invalidates relevant caches.
func (s *Service) Update(ctx context.Context, id string, u model.User) (WithRoles, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return WithRoles{}, fmt.Errorf("invalid uuid: %w", err)
	}
	u.ID = parsedID

	user, err := s.repo.Update(ctx, u)
	if err != nil {
		return WithRoles{}, err
	}

	// Invalidate caches
	s.cache.Delete(ctx, fmt.Sprintf("%s%s", userCacheKeyPrefix, id))
	s.cache.Delete(ctx, allUsersCacheKey)

	return user, nil
}

// Delete removes a user and invalidates relevant caches.
func (s *Service) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.cache.Delete(ctx, fmt.Sprintf("%s%s", userCacheKeyPrefix, id))
	s.cache.Delete(ctx, allUsersCacheKey)

	return nil
}
