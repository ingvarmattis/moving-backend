package interceptors

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ingvarmattis/moving/src/infra/utils"
)

var (
	errNoAuthTokenProvided = errors.New("no auth token provided")
	errInvalidAuthToken    = errors.New("invalid auth token")
)

func UnaryServerAuthInterceptor(validTokens []string) grpc.UnaryServerInterceptor {
	const (
		authKey      = "authorization"
		bearerPrefix = "Bearer "
	)

	validTokensMap := utils.ToMap(validTokens)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, errNoAuthTokenProvided.Error())
		}

		mdKey := md.Get(authKey)

		if len(mdKey) != 1 {
			return nil, status.Error(codes.Unauthenticated, errNoAuthTokenProvided.Error())
		}

		if !strings.HasPrefix(mdKey[0], bearerPrefix) {
			return nil, status.Error(codes.Unauthenticated, errInvalidAuthToken.Error())
		}

		token := strings.TrimPrefix(mdKey[0], bearerPrefix)
		if _, ok = validTokensMap[token]; !ok {
			return nil, status.Error(codes.Unauthenticated, errInvalidAuthToken.Error())
		}

		return handler(ctx, req)
	}
}
