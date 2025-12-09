package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (s *Server) withAuthorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			userIDValue string
			err         error
		)

		switch userIDValue, err = s.authenticate(req); err != nil {
		case errors.Is(err, errInvalidCookie):
			responseWithError(w, err, http.StatusBadRequest)
			return

		case errors.Is(err, http.ErrNoCookie),
			errors.Is(err, errInvalidAuthToken):
			userIDValue = uuid.New().String()
			authToken, err := s.newAuthToken(userIDValue)
			if err != nil {
				responseWithError(w, err, http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: userID.String(), Value: authToken})
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, userID, userIDValue)
		req = req.WithContext(ctx)

		h.ServeHTTP(w, req)
	})
}

func (s *Server) withAuthentification(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			userID      = "userID"
			userIDValue string
			err         error
		)

		switch userIDValue, err = s.authenticate(req); {
		case errors.Is(err, errInvalidCookie):
			responseWithError(w, err, http.StatusBadRequest)
			return

		case errors.Is(err, http.ErrNoCookie),
			errors.Is(err, errInvalidAuthToken):
			responseWithError(w, err, http.StatusUnauthorized)
			return
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, userID, userIDValue)
		req = req.WithContext(ctx)

		h.ServeHTTP(w, req)
	})
}

// authenticate extracts authCookie from the [http.Request], validates cookie, verifies its value
func (s *Server) authenticate(req *http.Request) (string, error) {
	authCookie, err := req.Cookie(userID.String())
	if err != nil {
		return "", err
	}

	if err = authCookie.Valid(); err != nil {
		return "", errors.Join(errInvalidCookie, err)
	}

	cutCookie, err := parseSignedCookie(authCookie.Value)
	if err != nil {
		return "", errors.Join(errInvalidAuthToken, errors.New("parsing auth token"))
	}

	if err = s.verify(cutCookie[0], authCookie.Value); err != nil {
		return "", err
	}

	return cutCookie[0], nil
}

func (s *Server) newAuthToken(uid string) (string, error) {
	signature, err := sign(uid, string(s.secretKey))
	if err != nil {
		return "", nil
	}
	sb := append([]byte(uid+separator), signature...)

	return hex.EncodeToString(sb), nil
}

type contextKey int

const userID contextKey = 1

func (c contextKey) String() string {
	if c == 1 {
		return "userID"
	}
	return ""
}

func (s *Server) verify(cutCookie, signedCookieValue string) error {
	signedCutCookieValue, err := s.newAuthToken(cutCookie)
	if err != nil {
		return errors.Join(errInvalidAuthToken, errors.New("invalid signature"))
	}

	if signedCutCookieValue != signedCookieValue {
		return errInvalidAuthToken
	}

	return nil
}

// parseSignedCookie decodes and cuts cookie into 2 parts. The first is the value, the second part is the signature of the value.
func parseSignedCookie(cookieValue string) (cookieParts []string, err error) {
	sb, err := hex.DecodeString(cookieValue)
	if err != nil {
		return nil, err
	}

	s := string(sb)
	cookieParts = strings.Split(s, separator)

	if len(cookieParts) != 2 {
		return nil, errInvalidAuthToken
	}

	return cookieParts, nil
}

func sign(s, secretKey string) ([]byte, error) {
	var (
		sb   = []byte(s)
		keyb = []byte(secretKey)
	)

	h := hmac.New(sha256.New, keyb)
	if _, err := h.Write(sb); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

const separator = "."

var (
	errInvalidCookie    error = errors.New("error invalid cookie")
	errInvalidAuthToken       = errors.New("error parsing signed cookie value")
)

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userID).(string)
	return userID, ok
}
