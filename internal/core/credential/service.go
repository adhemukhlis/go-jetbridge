package credential

import (
	"context"
	"fmt"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/proto/credential"
	"go-jetbridge/internal/infrastructure/cache"
	"go-jetbridge/internal/infrastructure/security"
	"go-jetbridge/internal/pkg/logger"
	"go-jetbridge/internal/provider/user"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

// Repository defines the interface for data access operations.
type Repository interface {
	FindByID(ctx context.Context, id string) (model.Credential, error)
	FindAll(ctx context.Context) ([]model.Credential, error)
	Create(ctx context.Context, c model.Credential) (model.Credential, error)
	Update(ctx context.Context, c model.Credential) (model.Credential, error)
	Delete(ctx context.Context, id string) error
}

// Service provides high-level business logic for Credential operations.
type Service struct {
	repo         Repository
	cache        cache.Cache[any]
	userProvider user.Provider
	sg           singleflight.Group
}

type ServiceOption func(*Service)

func WithCache(c cache.Cache[any]) ServiceOption {
	return func(s *Service) {
		s.cache = c
	}
}

func WithUserProvider(p user.Provider) ServiceOption {
	return func(s *Service) {
		s.userProvider = p
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
	credentialCacheKeyPrefix = "credential:"
	allCredentialsCacheKey   = "credentials:all"
)

func (s *Service) GetByID(ctx context.Context, id string) (*credential.CredentialResponse, error) {
	log := logger.FromCtx(ctx)
	cacheKey := credentialCacheKeyPrefix + id

	if s.cache != nil {
		if val, found := s.cache.Get(ctx, cacheKey); found {
			if cred, ok := val.(model.Credential); ok {
				log.Info("Cache Hit", "Credential ID", id)
				return mapCredentialToPB(cred), nil
			}
		}
	}

	log.Info("Cache Miss - Fetching from DB", "Credential ID", id)

	val, err, _ := s.sg.Do(cacheKey, func() (interface{}, error) {
		cred, dbErr := s.repo.FindByID(ctx, id)
		if dbErr != nil {
			return nil, dbErr
		}

		if s.cache != nil {
			s.cache.Set(ctx, cacheKey, cred, 10*time.Minute)
		}
		return cred, nil
	})

	if err != nil {
		return nil, err
	}

	cred := val.(model.Credential)
	return mapCredentialToPB(cred), nil
}

func (s *Service) GetAll(ctx context.Context) (*credential.CredentialListResponse, error) {
	log := logger.FromCtx(ctx)
	var creds []model.Credential

	if s.cache != nil {
		if val, found := s.cache.Get(ctx, allCredentialsCacheKey); found {
			if c, ok := val.([]model.Credential); ok {
				log.Info("Cache Hit", "All Credentials", true)
				creds = c
			}
		}
	}

	if creds == nil {
		log.Info("Cache Miss - Fetching all credentials from DB")

		val, err, _ := s.sg.Do(allCredentialsCacheKey, func() (interface{}, error) {
			dbCreds, dbErr := s.repo.FindAll(ctx)
			if dbErr != nil {
				return nil, dbErr
			}
			if s.cache != nil {
				s.cache.Set(ctx, allCredentialsCacheKey, dbCreds, 5*time.Minute)
			}
			return dbCreds, nil
		})

		if err != nil {
			return nil, err
		}
		creds = val.([]model.Credential)
	}

	var pbCreds []*credential.CredentialResponse
	for i := range creds {
		pbCreds = append(pbCreds, mapCredentialToPB(creds[i]))
	}

	return &credential.CredentialListResponse{
		Credentials: pbCreds,
	}, nil
}

func (s *Service) Create(ctx context.Context, req *credential.CreateCredentialRequest) (*credential.CredentialMinimumResponse, error) {
	// 1. Verify User ID via User Provider
	if s.userProvider != nil {
		_, err := s.userProvider.GetByID(ctx, req.UserId)
		if err != nil {
			return nil, fmt.Errorf("user not found or service error: %v", err)
		}
	}

	// 2. Hash Password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	c := model.Credential{
		UserId:               userID,
		PasswordHash:         hashedPassword,
		IsNeedPasswordChange: req.IsNeedPasswordChange != nil && *req.IsNeedPasswordChange,
	}

	if req.SuspendedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.SuspendedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid suspended_at: %w", err)
		}
		c.SuspendedAt = &t
	}
	if req.SuspendReason != nil {
		c.SuspendReason = req.SuspendReason
	}

	// 3. Save to Repo
	created, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, allCredentialsCacheKey)
	}

	return mapCredentialToMinimumPB(created.ID.String()), nil
}

func (s *Service) Update(ctx context.Context, req *credential.UpdateCredentialRequest) (*credential.CredentialMinimumResponse, error) {
	existing, err := s.repo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// 1. Verify User ID if changed
	if req.UserId != "" && req.UserId != existing.UserId.String() {
		if s.userProvider != nil {
			_, err := s.userProvider.GetByID(ctx, req.UserId)
			if err != nil {
				return nil, fmt.Errorf("user not found or service error: %v", err)
			}
		}
		userID, err := uuid.Parse(req.UserId)
		if err != nil {
			return nil, fmt.Errorf("invalid user id: %w", err)
		}
		existing.UserId = userID
	}

	// 2. Hash Password if provided
	if req.Password != "" {
		hashedPassword, err := security.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		existing.PasswordHash = hashedPassword
	}

	if req.IsNeedPasswordChange != nil {
		existing.IsNeedPasswordChange = *req.IsNeedPasswordChange
	}

	if req.SuspendedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.SuspendedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid suspended_at: %w", err)
		}
		existing.SuspendedAt = &t
	}

	if req.SuspendReason != nil {
		existing.SuspendReason = req.SuspendReason
	}

	// 3. Update in Repo
	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, credentialCacheKeyPrefix+req.Id)
		s.cache.Delete(ctx, allCredentialsCacheKey)
	}

	return mapCredentialToMinimumPB(updated.ID.String()), nil
}

func (s *Service) Delete(ctx context.Context, id string) (*credential.CredentialMinimumResponse, error) {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, credentialCacheKeyPrefix+id)
		s.cache.Delete(ctx, allCredentialsCacheKey)
	}

	return mapCredentialToMinimumPB(id), nil
}

func mapCredentialToPB(c model.Credential) *credential.CredentialResponse {
	return &credential.CredentialResponse{
		Id:                   c.ID.String(),
		UserId:               c.UserId.String(),
		IsNeedPasswordChange: c.IsNeedPasswordChange,
	}
}

func mapCredentialToMinimumPB(id string) *credential.CredentialMinimumResponse {
	return &credential.CredentialMinimumResponse{
		Id: id,
	}
}
