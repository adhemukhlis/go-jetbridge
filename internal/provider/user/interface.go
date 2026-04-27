package user

import (
	"context"
	userpb "go-jetbridge/internal/provider/user/gen/user"
)

// User represents the clean data structure consumed by the internal service logic.
// It is a type alias to the generated UserResponse to ensure it stays in sync with buf gen.
type User = userpb.UserResponse

// Provider defines the interface for interacting with the external User Service.
// It uses clean Go types to keep the service layer agnostic of gRPC details.
type Provider interface {
	GetByID(ctx context.Context, id string) (*User, error)
	Close() error
}
