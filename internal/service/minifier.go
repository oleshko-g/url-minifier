package service

import (
	"context"
	"io"

	"github.com/oleshko-g/url-minifier/internal/service/minifier"
)

// Minifier is the expected URL minifier service
//
//go:generate moq -pkg minifier -out ../mock/service/minifier.go . Minifier
type Minifier interface {
	MinifyURL(ctx context.Context, userID string, url string) (minifiedURL string, err error)
	MinifyURLs(ctx context.Context, userID string,
		urls []map[string]string) (minifiedURLs []map[string]string, err error)
	UnMinifyURL(id string) (url string, err error)
	UserURLs(ctx context.Context, userID string) ([]minifier.URL, error)
	// DeleteUserURLs deletes a batch of shortened URLs
	DeleteUserURLs(ctx context.Context, userID string, minifiedIDs []string) error
	UnMinifyUserURL(ctx context.Context, id string) (url string, isDeleted bool, err error)
	Ping() error
	io.Closer
	Stat
}

type Stat interface {
	TotalUsers(context.Context) (int, error)
	TotalURLs(context.Context) (int, error)
}
