package role

import (
	"context"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/proto/role"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	role.UnimplementedRoleServiceServer
	Service *Service
}

func (h *Handler) GetRole(ctx context.Context, req *role.GetRoleRequest) (*role.RoleResponse, error) {
	r, err := h.Service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return mapRoleToPB(r), nil
}

func (h *Handler) GetAllRole(ctx context.Context, _ *emptypb.Empty) (*role.RoleListResponse, error) {
	roles, err := h.Service.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var pbRoles []*role.RoleResponse
	for _, r := range roles {
		pbRoles = append(pbRoles, mapRoleToPB(r))
	}

	return &role.RoleListResponse{
		Roles: pbRoles,
	}, nil
}

func (h *Handler) CreateRole(ctx context.Context, req *role.CreateRoleRequest) (*role.RoleMinimumResponse, error) {
	m := model.Role{
		Key:  req.Key,
		Name: req.Name,
	}

	createdRole, err := h.Service.Create(ctx, m)
	if err != nil {
		return nil, err
	}

	return mapRoleToMinimumPB(createdRole.ID.String()), nil
}

func (h *Handler) UpdateRole(ctx context.Context, req *role.UpdateRoleRequest) (*role.RoleMinimumResponse, error) {
	existing, err := h.Service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	m := existing
	if req.Key != nil {
		m.Key = *req.Key
	}
	if req.Name != nil {
		m.Name = *req.Name
	}

	updatedRole, err := h.Service.Update(ctx, req.Id, m)
	if err != nil {
		return nil, err
	}

	return mapRoleToMinimumPB(updatedRole.ID.String()), nil
}

func (h *Handler) DeleteRole(ctx context.Context, req *role.DeleteRoleRequest) (*role.RoleMinimumResponse, error) {
	err := h.Service.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return mapRoleToMinimumPB(req.Id), nil
}

func mapRoleToPB(r model.Role) *role.RoleResponse {
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
