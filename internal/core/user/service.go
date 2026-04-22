// internal/core/user/service.go
package user

import (
	"context"
	"go-jetbridge/gen/jet/public/model"
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
}

// Service provides high-level business logic for User operations.
type Service struct {
	repo Repository
}

// NewService creates a new instance of Service with the given repository.
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetByID retrieves a single user with their roles by their ID.
func (s *Service) GetByID(ctx context.Context, id string) (WithRoles, error) {
	return s.repo.FindByID(ctx, id)
}

// GetAll retrieves all users with their roles.
func (s *Service) GetAll(ctx context.Context) ([]WithRoles, error) {
	return s.repo.FindAll(ctx)
}

// Create creates a new user.
func (s *Service) Create(ctx context.Context, u model.User) (WithRoles, error) {
	return s.repo.Create(ctx, u)
}
