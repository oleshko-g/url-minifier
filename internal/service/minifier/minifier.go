package minifier

import (
	"crypto/md5"
	"encoding/base64"
	"log"
)

type Service struct {
	storage Storager
	Config
}

type Storager interface {
	Save(key, value string) error
	Retrieve(key string) (value string, err error)
}

func New(s Storager, c *Config) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedID := encode([]byte(url), s.MaxLen)

	err = s.storage.Save(minifiedID, url)
	if err != nil {
		return "", err
	}

	minifiedURL = s.Config.BaseURL().String() + "/" + minifiedID

	log.Printf("minifiedURL: %s", minifiedURL)

	return minifiedURL, nil
}

func (s *Service) UnMinifyURL(id string) (url string, err error) {
	return s.storage.Retrieve(id)
}

func encode(data []byte, maxLen int) string {
	checksum := md5.Sum(data)
	return base64.RawURLEncoding.EncodeToString(checksum[:maxLen])
}
