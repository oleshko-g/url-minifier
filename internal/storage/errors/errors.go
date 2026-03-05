// Package errors is the internal storage package
package errors //revive:disable:var-naming

import (
	"errors"
)

var (
	// ErrNotFound is the error returned when a record is not found
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists is the error returned when a record already exists
	ErrAlreadyExists = errors.New("already exists")
	// ErrUnsupportedDataSource is the error returned when data source is not supported by the storage implementation
	ErrUnsupportedDataSource = errors.New("unsupported data source")
	// ErrEmptyParameter is the error returned when a mandatory parameter is empty
	ErrEmptyParameter = errors.New("empty parameter")
	// ErrAccessDenied is the error return when there's no permission to access data
	ErrAccessDenied = errors.New("access denied")
)
