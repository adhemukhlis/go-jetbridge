package credential

import (
	"context"
	"go-jetbridge/gen/proto/credential"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	credential.UnimplementedCredentialServiceServer
	Service *Service
}

func (h *Handler) GetCredential(ctx context.Context, req *credential.GetCredentialRequest) (*credential.CredentialResponse, error) {
	return h.Service.GetByID(ctx, req.Id)
}

func (h *Handler) GetAllCredential(ctx context.Context, _ *emptypb.Empty) (*credential.CredentialListResponse, error) {
	return h.Service.GetAll(ctx)
}

func (h *Handler) CreateCredential(ctx context.Context, req *credential.CreateCredentialRequest) (*credential.CredentialMinimumResponse, error) {
	return h.Service.Create(ctx, req)
}

func (h *Handler) UpdateCredential(ctx context.Context, req *credential.UpdateCredentialRequest) (*credential.CredentialMinimumResponse, error) {
	return h.Service.Update(ctx, req)
}

func (h *Handler) DeleteCredential(ctx context.Context, req *credential.DeleteCredentialRequest) (*credential.CredentialMinimumResponse, error) {
	return h.Service.Delete(ctx, req.Id)
}
