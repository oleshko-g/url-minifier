package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
)

var (
	ErrUnauthenticated     = errors.New(codes.Unauthenticated.String())
	ErrInternalServerError = errors.New(codes.Internal.String())
	ErrConflict            = errors.New(codes.AlreadyExists.String())
)
