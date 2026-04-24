package role

import (
	"context"
	"go-jetbridge/gen/proto/role"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	role.UnimplementedRoleServiceServer
	Service *Service
}

func (h *Handler) GetRole(ctx context.Context, req *role.GetRoleRequest) (*role.RoleResponse, error) {
	return h.Service.GetByID(ctx, req.Id)
}

func (h *Handler) GetAllRole(ctx context.Context, _ *emptypb.Empty) (*role.RoleListResponse, error) {
	return h.Service.GetAll(ctx)
}

func (h *Handler) CreateRole(ctx context.Context, req *role.CreateRoleRequest) (*role.RoleMinimumResponse, error) {
	return h.Service.Create(ctx, req)
}

func (h *Handler) UpdateRole(ctx context.Context, req *role.UpdateRoleRequest) (*role.RoleMinimumResponse, error) {
	return h.Service.Update(ctx, req)
}

func (h *Handler) DeleteRole(ctx context.Context, req *role.DeleteRoleRequest) (*role.RoleMinimumResponse, error) {
	return h.Service.Delete(ctx, req.Id)
}
