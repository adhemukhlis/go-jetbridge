package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buf.build/go/protovalidate"
)

// UnaryErrorInterceptor is a gRPC unary server interceptor that catches downstream errors
// and maps them to appropriate gRPC status codes, specifically handling database-related errors.
func UnaryErrorInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// If the error is already a gRPC status error, pass it through
	if _, ok := status.FromError(err); ok {
		return resp, err
	}

	// Handle validation errors from protovalidate and map to Google's errdetails.BadRequest
	var valErr *protovalidate.ValidationError
	if errors.As(err, &valErr) {
		st := status.New(codes.InvalidArgument, "validation failed")
		br := &errdetails.BadRequest{}

		for _, v := range valErr.Violations {
			var fieldName string
			if field := v.Proto.GetField(); field != nil {
				elements := field.GetElements()
				if len(elements) > 0 {
					fieldName = elements[len(elements)-1].GetFieldName()
				}
			}

			br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       fieldName,
				Description: v.Proto.GetMessage(),
			})
		}

		stWithDetails, attachErr := st.WithDetails(br)
		if attachErr == nil {
			return nil, stWithDetails.Err()
		}
		// Fallback if WithDetails fails
		return nil, st.Err()
	}

	// Map specific database errors using PostgreSQL codes
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		fmt.Printf("DEBUG PG ERROR: code=%s, message=%s\n", pgErr.Code, pgErr.Message)

		switch pgErr.Code {
		case "23505":
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		case "22P02":
			return nil, status.Error(codes.NotFound, "resource not found")
		}
	}

	// Map common "no rows" errors to NotFound status
	if errors.Is(err, qrm.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	// Default to Internal error for any unhandled application-level errors
	return nil, status.Errorf(codes.Internal, "internal error: %v", err)
}
