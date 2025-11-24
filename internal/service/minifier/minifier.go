package minifier

import (
	"crypto/md5"
	"encoding/base64"
	"errors"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// Service is the implementation of [http.Service]
type Service struct {
	storage Storager
	*Config
}

// Storager is the expected implementation of storage for the URL minifier
//
//go:generate moq -pkg file -out ../../mock/storage/storage.go . Storager
type Storager interface {
	Save(key, value string) error
	SaveList(values []map[string]string) error
	Retrieve(key string) (value string, err error)
	Ping() error
}

// New configures a URL minifier service with the passed [Storager] and [Config]
func New(s Storager, cp *Config) *Service {
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
func (s *Service) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedID := encode([]byte(url), s.MaxLen)

	err = s.storage.Save(minifiedID, url)
	if err != nil {
		if !errors.Is(err, storageErrors.ErrAlreadyExists) {
			return "", err
		}
		err = ErrMinifiedAlready
	}

	minifiedURL = s.Config.BaseURL().String() + "/" + minifiedID

	return minifiedURL, err
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

// MinifyURLs takes a slice of URLs coupled with their correlation ids an return the slice of minified URLs coupled with correlation ids or the first error occurred
//
// TODO: add tests
// TODO: rewrite to use storage.SaveList
func (s *Service) MinifyURLs(urls []map[string]string) (minifiedURLs []map[string]string, err error) {
	for _, url := range urls {
		minifiedURL := make(map[string]string, 1)
		for correlationID, originalURL := range url {
			minifiedURL[correlationID], err = s.MinifyURL(originalURL)
		}

		if err != nil {
			if !errors.Is(err, ErrMinifiedAlready) {
				return nil, err
			}
			// keep  err == [ErrMinifiedAlready] untill the end or encounterig a real error
		}
		minifiedURLs = append(minifiedURLs, minifiedURL)
	}
	return minifiedURLs, err
}
