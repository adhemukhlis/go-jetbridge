// Package middleware provides gRPC interceptors for cross-cutting concerns such as authentication.
package middleware

import (
	"context"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// apiKeyHeader is the metadata key used to transmit the API key for authentication.
const apiKeyHeader = "x-api-key"

// UnaryAuthInterceptor is a gRPC unary server interceptor that validates the API Key
// transmitted in the request metadata. It compares the provided key against the
// API_KEY environment variable. If the key is missing or invalid, it returns
// an Unauthenticated error.
func UnaryAuthInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract metadata from the incoming request context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	// Retrieve the API key from the "x-api-key" header
	values := md.Get(apiKeyHeader)
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "API Key is required")
	}

	// Compare the provided key with the expected key from the environment variable
	providedKey := values[0]
	expectedKey := os.Getenv("API_KEY")

	if providedKey != expectedKey {
		return nil, status.Errorf(codes.Unauthenticated, "invalid API Key")
	}

	// If authentication succeeds, proceed to the actual handler
	return handler(ctx, req)
}
