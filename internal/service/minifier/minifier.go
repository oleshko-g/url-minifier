package minifier

import (
	"crypto/md5"
	"encoding/base64"
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
	Retrieve(key string) (value string, err error)
}

// New configures a URL minifier service with the passed [Storager] and [Config]
func New(s Storager, cp *Config) *Service {
	return &Service{
		storage: s,
		Config:  cp,
	}
}

// MinifyURL takes any string encodes it and returns the minified URL or an error
//
// TODO: add tests
func (s *Service) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedID := encode([]byte(url), s.MaxLen)

	err = s.storage.Save(minifiedID, url)
	if err != nil {
		return "", err
	}

	minifiedURL = s.Config.BaseURL().String() + "/" + minifiedID

	return minifiedURL, nil
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
