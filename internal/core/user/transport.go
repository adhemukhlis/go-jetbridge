package user

import (
	"context"
	"go-jetbridge/gen/proto/user"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	user.UnimplementedUserServiceServer
	Service *Service
}

func (h *Handler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	return h.Service.GetByID(ctx, req.Id)
}

func (h *Handler) GetAllUser(ctx context.Context, _ *emptypb.Empty) (*user.UserListResponse, error) {
	return h.Service.GetAll(ctx)
}

func (h *Handler) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserMinimumResponse, error) {
	return h.Service.Create(ctx, req)
}

func (h *Handler) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserMinimumResponse, error) {
	return h.Service.Update(ctx, req)
}

func (h *Handler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.UserMinimumResponse, error) {
	return h.Service.Delete(ctx, req.Id)
}
