package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/oleshko-g/url-minifier/internal/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// authOption checks if ctx has "authorization" metadata.
//   - If ctx has "authorization" authOption authenticates "authorization" first value.
//   - If ctx has no "authorization" authOption puts "authorization" is ctx.
func (s *Server) authOption() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req any, srvInfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var (
			userID string
			md     metadata.MD
		)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(map[string]string{})
		}

		authVals := md.Get("auhorization")
		var authTokenValue string
		if len(authVals) > 0 {
			authTokenValue = authVals[0]
		} else {
			return nil, ErrUnauthenticated
		}

		userID, err := token.Authenticate(authTokenValue, s.Config.SecretKey.String())
		if err != nil {
			if errors.Is(err, token.ErrInvalidSignature) {
				return nil, ErrUnauthenticated
			}

			if errors.Is(err, token.ErrInvalidAuthTokenValue) {
				id, _ := uuid.NewV7()
				userID = id.String()
				signedUserID, err := token.New(userID, s.Config.SecretKey.String())
				if err != nil {
					return nil, err
				}
				md.Set("authorization", signedUserID)
			}
		}

		md.Set("userID", userID)

		return handler(ctx, req)
	})
}
