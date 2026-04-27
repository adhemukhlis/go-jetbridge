package user

import (
	"context"
	"go-jetbridge/internal/provider/user/gen/user"

	"google.golang.org/grpc"
)

type providerImpl struct {
	client user.UserServiceClient
	conn   *grpc.ClientConn
}

// NewProvider creates a new instance of User Provider.
func NewProvider() (Provider, error) {
	conn, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &providerImpl{
		client: user.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (p *providerImpl) GetByID(ctx context.Context, id string) (*User, error) {
	resp, err := p.client.GetUser(ctx, &user.GetUserRequest{Id: id})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *providerImpl) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
