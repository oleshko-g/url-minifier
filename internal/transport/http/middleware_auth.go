package http //revive:disable-line:var-naming

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/oleshko-g/url-minifier/internal/token"
)

func (s *Server) withAuthorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			userIDValue string
			err         error
		)

		switch userIDValue, err = s.authenticate(req); {
		case errors.Is(err, token.ErrInvalidSignature):
			responseWithError(w, err, http.StatusUnauthorized)
			return
		case errors.Is(err, errInvalidCookie):
			responseWithError(w, err, http.StatusBadRequest)
			return

		case errors.Is(err, http.ErrNoCookie),
			errors.Is(err, token.ErrInvalidAuthTokenValue):
			userIDValue = uuid.New().String()
			authToken, err := token.New(userIDValue, s.SecretKey.String())
			if err != nil {
				responseWithError(w, err, http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: contextKeyUserID.String(), Value: authToken})
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, contextKeyUserID, userIDValue)
		req = req.WithContext(ctx)

		h.ServeHTTP(w, req)
	})
}

// authenticate extracts authCookie from the [http.Request], validates cookie, verifies its value
func (s *Server) authenticate(req *http.Request) (string, error) {
	authCookie, err := req.Cookie(contextKeyUserID.String())
	if err != nil {
		return "", err
	}

	if err = authCookie.Valid(); err != nil {
		return "", errors.Join(errInvalidCookie, err)
	}

	cutCookie, err := token.ParseSignedToken(authCookie.Value)
	if err != nil {
		return "", err
	}

	if err = token.Verify(cutCookie[0], authCookie.Value, s.SecretKey.String()); err != nil {
		return "", err
	}

	return cutCookie[0], nil
}

type contextKey int

const contextKeyUserID contextKey = 1

func (c contextKey) String() string {
	if c == 1 {
		return "userID"
	}
	return ""
}

var (
	errInvalidCookie    error = errors.New("error invalid cookie")
)

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok
}

func (s *Server) authorized(h handlerWithUserID) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var err error
		ctx := req.Context()
		uid, ok := userIDFromContext(ctx)
		if !ok {
			err = errors.New("no userID in the request")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}
		h(uid, res, req)
	})
}
