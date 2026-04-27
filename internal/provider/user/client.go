package user

import (
	"context"
	"fmt"
	"go-jetbridge/internal/pkg/config"
	"go-jetbridge/internal/pkg/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClient initializes a new gRPC connection to the User Service.
func NewClient() (*grpc.ClientConn, error) {
	host := config.GetEnv("USER_SERVICE_HOST", "localhost")
	port := config.GetEnvInt("USER_SERVICE_PORT", 50052)
	address := fmt.Sprintf("%s:%d", host, port)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(loggingInterceptor),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service at %s: %w", address, err)
	}

	return conn, nil
}

func loggingInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()
	log := logger.FromCtx(ctx).With("grpc.method", method, "grpc.target", cc.Target())

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(start)
	if err != nil {
		log.Error("gRPC call failed", "duration", duration, "error", err)
	} else {
		log.Info("gRPC call success", "duration", duration)
	}

	return err
}
