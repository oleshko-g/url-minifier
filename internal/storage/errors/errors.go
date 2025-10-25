package errors

import (
	"errors"
)

var (
	// ErrNotFound is the error returned when a record is not found
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists is the error return whena a recod already exists
	ErrAlreadyExists = errors.New("already exists")
)
