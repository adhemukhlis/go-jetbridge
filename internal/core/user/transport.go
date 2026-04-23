// internal/core/user/transport.go
package user

import (
	"context"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/proto/role"
	"go-jetbridge/gen/proto/user"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	user.UnimplementedUserServiceServer
	Service *Service
}

func (h *Handler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	u, err := h.Service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return mapUserToPB(u), nil
}

func (h *Handler) GetAllUser(ctx context.Context, _ *emptypb.Empty) (*user.UserListResponse, error) {
	users, err := h.Service.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var pbUsers []*user.UserResponse
	for _, u := range users {
		pbUsers = append(pbUsers, mapUserToPB(u))
	}

	return &user.UserListResponse{
		Users: pbUsers,
	}, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserMinimumResponse, error) {
	u := model.User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
	}

	createdUser, err := h.Service.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	return mapUserToMinimumPB(createdUser.ID.String()), nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserMinimumResponse, error) {
	// For partial updates, we could fetch then merge, or build a dynamic update.
	// For "CRUD biasa", we'll fetch then merge here to keep repository simple.
	existing, err := h.Service.GetByID(ctx, req.Id)
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

	updatedUser, err := h.Service.Update(ctx, req.Id, u)
	if err != nil {
		return nil, err
	}

	return mapUserToMinimumPB(updatedUser.ID.String()), nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.UserMinimumResponse, error) {
	err := h.Service.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return mapUserToMinimumPB(req.Id), nil
}

// mapUserToPB translates the internal user model (with roles) to the Protobuf response format.
func mapUserToPB(u WithRoles) *user.UserResponse {
	var pbRoles []*role.RoleResponse
	for _, r := range u.Role {
		pbRoles = append(pbRoles, &role.RoleResponse{
			Id:   r.ID.String(),
			Key:  r.Key,
			Name: r.Name,
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
