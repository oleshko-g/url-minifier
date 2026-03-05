package minifier

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"net/url"

	"github.com/oleshko-g/url-minifier/internal/storage"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// Service is the implementation of [http.Service]
type Service struct {
	storage storage.Storager
	*Config
}

// New configures a URL minifier service with the passed [Storager] and [Config]
func New(s storage.Storager, cp *Config) *Service {
	return &Service{
		storage: s,
		Config:  cp,
	}
}

// Ping check if the storage is up
func (s *Service) Ping() error {
	return s.storage.Ping()
}

// MinifyURL takes any string, encodes it and returns the minified URL or an error. If the original URL is minified already MinifyURL returns both non empty minifiedURL and [ErrMinifiedAlready] error
//
// TODO: add tests
func (s *Service) MinifyURL(ctx context.Context, userID, originalURL string) (minifiedURL string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	minifiedID := encode([]byte(originalURL), s.MaxLen)
	minifiedURL = s.newMinifiedURL(minifiedID)

	err = s.storage.SaveUserString(ctx, storage.UserString{
		UserID: userID,
		Key:    minifiedID,
		Value:  originalURL,
	})
	if err != nil {
		if !errors.Is(err, storageErrors.ErrAlreadyExists) {
			return "", err
		}
		err = ErrMinifiedAlready
	}

	return minifiedURL, err
}

func (s *Service) newMinifiedURL(minifiedID string) string {
	return s.Config.BaseURL().String() + "/" + minifiedID
}

// UnMinifyURL takes an id of the minified URL and returns the stored original URL or an error
//
// TODO: add tests
func (s *Service) UnMinifyURL(id string) (url string, err error) {
	return s.storage.Retrieve(id)
}

// TODO: research if truncation might lead to collisions
func encode(data []byte, maxLen int) string {
	checksum := md5.Sum(data)
	return base64.RawURLEncoding.EncodeToString(checksum[:maxLen])
}

// MinifyURLs takes a slice of URLs coupled with their correlation ids.
// It returns a slice of minified URLs coupled with correlation ids or the first error occurred
//
// a map key is a correlation ID
// "urls" map value is an original URL
// "minifiedURLs" map value is a minifiedURL
// TODO: add tests
// TODO: rewrite to use storage.SaveList
func (s *Service) MinifyURLs(ctx context.Context, userID string, originalURLs []map[string]string) (minifiedURLs []map[string]string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	for _, oURL := range originalURLs {
		minifiedURL := make(map[string]string, 1)
		for correlationID, originalURL := range oURL {
			minifiedURL[correlationID], err = s.MinifyURL(ctx, userID, originalURL)
		}

		if err != nil {
			if !errors.Is(err, ErrMinifiedAlready) {
				return nil, err
			}
			// keep  err == [ErrMinifiedAlready] until the end or encountering a real error
		}
		minifiedURLs = append(minifiedURLs, minifiedURL)
	}
	return minifiedURLs, err
}

// UserURLs takes userID and return a slice of [minifier.URL]'s or an error
func (s *Service) UserURLs(ctx context.Context, userID string) ([]URL, error) {
	uss, err := s.storage.RetrieveUserStrings(ctx, userID)
	if err != nil {
		return nil, err
	}

	var urls []URL
	for _, us := range uss {
		mURL := s.newURL(us.Key, us.Value)

		urls = append(urls, mURL)
	}
	return urls, nil
}

func (s *Service) newURL(id, originalURL string) URL {
	oURL, _ := url.Parse(originalURL)
	mURL, _ := url.Parse(s.newMinifiedURL(id))
	return URL{
		OriginalURL: oURL,
		MinifiedURL: mURL,
	}
}

// DeleteUserURLs takes userID and a slice of minified IDs and marks as deleted the associated minified URLs
func (s *Service) DeleteUserURLs(userID string, minifiedIDs []string) error {
	ctx := context.Background()
	var successCh = make(chan struct{}, len(minifiedIDs))
	var errCh = make(chan error, len(minifiedIDs))
	for _, mID := range minifiedIDs {
		go func() {
			errCh <- s.storage.MarkDeletedUserString(ctx, userID, mID)
			successCh <- struct{}{}
		}()
	}

	for i := 0; i < len(minifiedIDs); i++ {
		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-successCh:
		}
	}
	return nil
}

// UnMinifyUserURL takes an id of a user string and returned its value and if it's deleted
func (s *Service) UnMinifyUserURL(ctx context.Context, id string) (value string, isDeleted bool, err error) {
	dus, err := s.storage.RetrieveUserString(ctx, id)
	if err != nil {
		return "", false, err
	}

	return dus.Value, dus.Deleted, nil
}
