package role

import (
	"context"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/proto/role"
	"go-jetbridge/internal/infrastructure/cache"
	"go-jetbridge/internal/pkg/logger"
	"time"

	"golang.org/x/sync/singleflight"
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
	sg    singleflight.Group
}

type ServiceOption func(*Service)

func WithCache(c cache.Cache[any]) ServiceOption {
	return func(s *Service) {
		s.cache = c
	}
}

func NewService(repo Repository, opts ...ServiceOption) *Service {
	s := &Service{
		repo: repo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

const (
	roleCacheKeyPrefix = "role:"
	allRolesCacheKey   = "roles:all"
)

func (s *Service) GetByID(ctx context.Context, id string) (*role.RoleResponse, error) {
	log := logger.FromCtx(ctx)
	cacheKey := roleCacheKeyPrefix + id

	if s.cache != nil {
		if val, found := s.cache.Get(ctx, cacheKey); found {
			if r, ok := val.(model.Role); ok {
				log.Info("Cache Hit", "Role ID", id)
				return mapRoleToPB(&r), nil
			}
		}
	}

	log.Info("Cache Miss - Fetching from DB", "Role ID", id)

	val, err, _ := s.sg.Do(cacheKey, func() (interface{}, error) {
		r, dbErr := s.repo.FindByID(ctx, id)
		if dbErr != nil {
			return nil, dbErr
		}

		if s.cache != nil {
			s.cache.Set(ctx, cacheKey, r, 10*time.Minute)
		}
		return r, nil
	})

	if err != nil {
		return nil, err
	}

	r := val.(model.Role)
	return mapRoleToPB(&r), nil
}

func (s *Service) GetAll(ctx context.Context) (*role.RoleListResponse, error) {
	log := logger.FromCtx(ctx)
	var roles []model.Role

	if s.cache != nil {
		if val, found := s.cache.Get(ctx, allRolesCacheKey); found {
			if r, ok := val.([]model.Role); ok {
				log.Info("Cache Hit", "All Roles", true)
				roles = r
			}
		}
	}

	if roles == nil {
		log.Info("Cache Miss - Fetching all roles from DB")

		val, err, _ := s.sg.Do(allRolesCacheKey, func() (interface{}, error) {
			dbRoles, dbErr := s.repo.FindAll(ctx)
			if dbErr != nil {
				return nil, dbErr
			}
			if s.cache != nil {
				s.cache.Set(ctx, allRolesCacheKey, dbRoles, 5*time.Minute)
			}
			return dbRoles, nil
		})

		if err != nil {
			return nil, err
		}
		roles = val.([]model.Role)
	}

	var pbRoles []*role.RoleResponse
	for i := range roles {
		pbRoles = append(pbRoles, mapRoleToPB(&roles[i]))
	}

	return &role.RoleListResponse{
		Roles: pbRoles,
	}, nil
}

func (s *Service) Create(ctx context.Context, req *role.CreateRoleRequest) (*role.RoleMinimumResponse, error) {
	m := model.Role{
		Key:  req.Key,
		Name: req.Name,
	}

	createdRole, err := s.repo.Create(ctx, m)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, allRolesCacheKey)
	}

	return mapRoleToMinimumPB(createdRole.ID.String()), nil
}

func (s *Service) Update(ctx context.Context, req *role.UpdateRoleRequest) (*role.RoleMinimumResponse, error) {
	existing, err := s.repo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if req.Key != nil {
		existing.Key = *req.Key
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}

	updatedRole, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, roleCacheKeyPrefix+req.Id)
		s.cache.Delete(ctx, allRolesCacheKey)
	}

	return mapRoleToMinimumPB(updatedRole.ID.String()), nil
}

func (s *Service) Delete(ctx context.Context, id string) (*role.RoleMinimumResponse, error) {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, roleCacheKeyPrefix+id)
		s.cache.Delete(ctx, allRolesCacheKey)
	}

	return mapRoleToMinimumPB(id), nil
}

func mapRoleToPB(r *model.Role) *role.RoleResponse {
	return &role.RoleResponse{
		Id:   r.ID.String(),
		Key:  r.Key,
		Name: r.Name,
	}
}

func mapRoleToMinimumPB(id string) *role.RoleMinimumResponse {
	return &role.RoleMinimumResponse{
		Id: id,
	}
}
