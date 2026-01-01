package interceptor

import (
	"context"
	"strings"

	"github.com/eduardovfaleiro/gatekeeper/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is required")
		}

		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

		userID, err := token.ValidateToken(tokenStr, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		newCtx := context.WithValue(ctx, "user_id", userID)

		return handler(newCtx, req)
	}
}

var publicMethods = map[string]struct{}{
	"/auth.AuthService/Login":    {},
	"/auth.AuthService/Register": {},
}

func isPublicMethod(method string) bool {
	_, ok := publicMethods[method]
	return ok
}
