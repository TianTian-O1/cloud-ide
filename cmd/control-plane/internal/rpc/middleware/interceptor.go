package middleware

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

var RecoveredErr = errors.New("recovered")

// RecoveryInterceptorMiddleware 防止panic导致整个服务崩溃
func RecoveryInterceptorMiddleware(logger *logr.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(RecoveredErr, "", "info", err)
			}
		}()

		return handler(ctx, req)
	}
}
