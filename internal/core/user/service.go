package user

import (
	"context"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/proto/role"
	"go-jetbridge/gen/proto/user"
	"go-jetbridge/internal/infrastructure/cache"
	"go-jetbridge/internal/pkg/logger"
	"time"

	"golang.org/x/sync/singleflight"
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
	sg    singleflight.Group
}

type ServiceOption func(*Service)

func WithCache(c cache.Cache[any]) ServiceOption {
	return func(s *Service) {
		s.cache = c
	}
}

// NewService creates a new instance of Service with the given repository and options.
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
	userCacheKeyPrefix = "user:"
	allUsersCacheKey   = "users:all"
)

// GetByID retrieves a single user with their roles by their ID.
func (s *Service) GetByID(ctx context.Context, id string) (*user.UserResponse, error) {
	log := logger.FromCtx(ctx)
	cacheKey := userCacheKeyPrefix + id

	// Try to get from cache
	if s.cache != nil {
		if val, found := s.cache.Get(ctx, cacheKey); found {
			if u, ok := val.(WithRoles); ok {
				log.Info("Cache Hit", "User ID", id)
				return mapUserToPB(&u), nil
			}
		}
	}

	log.Info("Cache Miss - Fetching from DB", "User ID", id)

	// Singleflight: protect against cache stampede
	val, err, _ := s.sg.Do(cacheKey, func() (interface{}, error) {
		u, dbErr := s.repo.FindByID(ctx, id)
		if dbErr != nil {
			return nil, dbErr
		}

		if s.cache != nil {
			s.cache.Set(ctx, cacheKey, u, 10*time.Minute)
		}
		return u, nil
	})

	if err != nil {
		return nil, err
	}

	u := val.(WithRoles)
	return mapUserToPB(&u), nil
}

// GetAll retrieves all users with their roles.
func (s *Service) GetAll(ctx context.Context) (*user.UserListResponse, error) {
	log := logger.FromCtx(ctx)
	var users []WithRoles

	// Try to get from cache
	if s.cache != nil {
		if val, found := s.cache.Get(ctx, allUsersCacheKey); found {
			if u, ok := val.([]WithRoles); ok {
				log.Info("Cache Hit", "All Users", true)
				users = u
			}
		}
	}

	if users == nil {
		log.Info("Cache Miss - Fetching all users from DB")

		// Singleflight
		val, err, _ := s.sg.Do(allUsersCacheKey, func() (interface{}, error) {
			dbUsers, dbErr := s.repo.FindAll(ctx)
			if dbErr != nil {
				return nil, dbErr
			}
			if s.cache != nil {
				s.cache.Set(ctx, allUsersCacheKey, dbUsers, 5*time.Minute)
			}
			return dbUsers, nil
		})

		if err != nil {
			return nil, err
		}
		users = val.([]WithRoles)
	}

	var pbUsers []*user.UserResponse
	// Zero-copy looping
	for i := range users {
		pbUsers = append(pbUsers, mapUserToPB(&users[i]))
	}

	return &user.UserListResponse{
		Users: pbUsers,
	}, nil
}

// Create creates a new user and invalidates the "all users" cache.
func (s *Service) Create(ctx context.Context, req *user.CreateUserRequest) (*user.UserMinimumResponse, error) {
	u := model.User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
	}

	createdUser, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	// Invalidate the list cache
	if s.cache != nil {
		s.cache.Delete(ctx, allUsersCacheKey)
	}

	return mapUserToMinimumPB(createdUser.ID.String()), nil
}

// Update updates an existing user and invalidates relevant caches.
func (s *Service) Update(ctx context.Context, req *user.UpdateUserRequest) (*user.UserMinimumResponse, error) {
	existing, err := s.repo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	u := existing.User
	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Username != nil {
		u.Username = *req.Username
	}
	if req.Email != nil {
		u.Email = *req.Email
	}

	updatedUser, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	if s.cache != nil {
		s.cache.Delete(ctx, userCacheKeyPrefix+req.Id)
		s.cache.Delete(ctx, allUsersCacheKey)
	}

	return mapUserToMinimumPB(updatedUser.ID.String()), nil
}

// Delete removes a user and invalidates relevant caches.
func (s *Service) Delete(ctx context.Context, id string) (*user.UserMinimumResponse, error) {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	if s.cache != nil {
		s.cache.Delete(ctx, userCacheKeyPrefix+id)
		s.cache.Delete(ctx, allUsersCacheKey)
	}

	return mapUserToMinimumPB(id), nil
}

// mapUserToPB takes a pointer to avoid struct copying overhead
func mapUserToPB(u *WithRoles) *user.UserResponse {
	var pbRoles []*role.RoleResponse
	for i := range u.Role {
		pbRoles = append(pbRoles, &role.RoleResponse{
			Id:   u.Role[i].ID.String(),
			Key:  u.Role[i].Key,
			Name: u.Role[i].Name,
		})
	}

	return &user.UserResponse{
		Id:       u.ID.String(),
		Name:     u.Name,
		Username: u.Username,
		Email:    u.Email,
		Roles:    pbRoles,
	}
}

func mapUserToMinimumPB(id string) *user.UserMinimumResponse {
	return &user.UserMinimumResponse{
		Id: id,
	}
}
