// Package auth is the package to authorize and authenticate users
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

func NewAuthToken(userID, secret string) (string, error) {
	signature, err := sign(userID, secret)
	if err != nil {
		return "", nil
	}

	sb := strings.Builder{}
	sb.WriteString(userID)
	sb.WriteString(separator)
	sb.Write(signature)

	return hex.EncodeToString([]byte(sb.String())), nil
}

type contextKey int

const separator = "."

func sign(s, secretKey string) ([]byte, error) {
	var (
		sb  = []byte(s)
		key = []byte(secretKey)
	)

	h := hmac.New(sha256.New, key)
	if _, err := h.Write(sb); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func verify(tokenToVerify, signedTokenValue, secret string) error {
	signedTokenToVerify, err := NewAuthToken(tokenToVerify, secret)
	if err != nil {
		return errors.Join(ErrInvalidAuthToken, ErrInvalidSignature)
	}

	if signedTokenToVerify != signedTokenValue {
		return ErrInvalidAuthToken
	}

	return nil
}

// parseSignedToken decodes and cuts cookie into 2 parts. The first is the value, the second part is the signature of the value.
func parseSignedToken(tokenValue string) (tokenParts []string, err error) {
	sb, err := hex.DecodeString(tokenValue)
	if err != nil {
		return nil, err
	}

	s := string(sb)
	tokenParts = strings.Split(s, separator)

	if len(tokenParts) != 2 {
		return nil, ErrInvalidAuthToken
	}

	return tokenParts, nil
}

var (
	ErrInvalidAuthToken = errors.New("error parsing signed cookie value")
	ErrInvalidCookie    = errors.New("error invalid cookie")
	ErrInvalidSignature = errors.New("invalid signature")
)
