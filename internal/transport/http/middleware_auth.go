package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) withAuthorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			authToken   *http.Cookie
			userIDKey   string = "userID"
			userIDValue string
		)

		switch err := authorized(userIDKey, req); err != nil {
		case errors.Is(err, errInvalidCookie):
			responseWithError(w, err, http.StatusBadRequest)
			return
		case errors.Is(err, http.ErrNoCookie), errors.Is(err, errInvalidAuthToken):
			userIDValue = uuid.New().String()
			if authToken, err = s.newSignedCookie(userIDKey, userIDValue); err != nil {
				responseWithError(w, err, http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, authToken)
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userIDValue)
		req = req.WithContext(ctx)

		h.ServeHTTP(w, req)
	})
}

func authorized(authCookieName string, req *http.Request) error {
	c, err := req.Cookie(authCookieName)
	if err != nil {
		return err
	}

	if err = c.Valid(); err != nil {
		return errors.Join(errInvalidCookie, err)
	}

	if err = validAuthToken(c.Value); err != nil {
		return err
	}

	return nil
}
func validAuthToken(value string) error {
	var authToken string
	authToken, err := parseAuthToken(value)
	if err != nil {
		return errors.Join(errInvalidAuthToken, errors.New("parsing auth token"))
	}

	if !validateSignature(authToken) {
		return errors.Join(errInvalidAuthToken, errors.New("invalid signature"))
	}
	return nil
}

// TODO: decode hex, validateSignature
func parseAuthToken(cookieValue string) (string, error) {
	return "", nil
}

// TODO: re-sign, compare
func validateSignature(signedString string) bool {
	return true
}

func (s *Server) newSignedCookie(userIDKey, userID string) (*http.Cookie, error) {
	signature, err := sign(userID, string(s.secretKey))
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:  userIDKey,
		Value: signature,
	}, nil
}

// TODO: sign
func sign(s, secretKey string) (string, error) {
	h := hmac.New(sha256.New, []byte(secretKey))
	if _, err := h.Write([]byte(s)); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

var (
	errInvalidCookie    error = errors.New("error invalid cookie")
	errNoUserIDCookie         = errors.New("error no userID")
	_                         = errNoUserIDCookie
	errInvalidAuthToken       = errors.New("error invalid auth token")
)
