package middleware

import (
	"context"
	"log/slog"
	"time"

	"go-jetbridge/internal/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryLoggerInterceptor logs the gRPC request details using slog and injects trace_id.
func UnaryLoggerInterceptor(baseLogger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Extract trace_id from metadata or generate a new UUIDv7
		traceID := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-trace-id"); len(vals) > 0 {
				traceID = vals[0]
			}
		}

		if traceID == "" {
			newID, _ := uuid.NewV7()
			traceID = newID.String()
		}

		// Create a child logger with trace_id
		ctxLogger := baseLogger.With("trace_id", traceID)

		// Inject logger into context for downstream usage
		ctx = logger.WithCtx(ctx, ctxLogger)

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Internal
			}
			ctxLogger.Error("gRPC request failed",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
				slog.String("duration", duration.String()),
				slog.Any("error", err),
			)
		} else {
			ctxLogger.Info("gRPC request success",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
				slog.String("duration", duration.String()),
			)
		}

		return resp, err
	}
}
