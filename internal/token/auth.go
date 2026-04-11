// Package token is the package to [Sign], [Authenticate], [Authorize] tokens.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// Authenticate parses the signed token value, verifies the signature, and returns the token value or an error.
func Authenticate(signedTokenValue, secret string) (string, error) {
	tokenParts, err := parseSignedToken(signedTokenValue)
	if err != nil {
		return "", err
	}
	err = verify(tokenParts[0], signedTokenValue, secret)
	if err != nil {
		return "", err
	}

	return tokenParts[0], nil
}

// New signs the token with the secret and returns the signed token value encoded hex or an error.
func New(token, secret string) (signedToken string, err error) {
	signature, err := sign(token, secret)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	sb.WriteString(token)
	sb.WriteString(separator)
	sb.Write(signature)

	return hex.EncodeToString([]byte(sb.String())), nil
}

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

// verify re-signs tokenValue with the secret and compares against signedTokenValue.
// If resignedTokenValue isn't equal to signedTokenValue, returns [ErrInvalidAuthTokenValue].
func verify(tokenValue, signedTokenValue, secret string) error {
	resignedTokenValue, err := New(tokenValue, secret)
	if err != nil {
		return err
	}

	if resignedTokenValue != signedTokenValue {
		return ErrInvalidSignature
	}

	return nil
}

// parseSignedToken decodes and cuts tokenValue into 2 parts.
// The 1st part is the value.
// The 2nd part is the signature of the value.
func parseSignedToken(tokenValue string) (tokenParts [2]string, err error) {
	decodedTokenValue, err := hex.DecodeString(tokenValue)
	if err != nil {
		return [2]string{}, err
	}

	tokenParts = [2]string(strings.Split(string(decodedTokenValue), separator)[:2])

	if len(tokenParts) != 2 {
		return [2]string{}, ErrInvalidAuthTokenValue
	}

	return tokenParts, nil
}

var (
	ErrInvalidAuthTokenValue = errors.New("error parsing signed token value")
	ErrInvalidSignature      = errors.New("error invalid signature")
)
