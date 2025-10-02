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

func New(s Storager) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedURL = s.baseURL.String() + "/" + encode([]byte(url), s.MaxLen)
	log.Printf("minifiedURL: %s", minifiedURL)
	err = s.storage.Save(encode([]byte(url), s.MaxLen), url)
	if err != nil {
		return "", err
	}
	return minifiedURL, nil
}

func (s *Service) UnMinifyURL(id string) (url string, err error) {
	return s.storage.Retrieve(id)
}

func encode(data []byte, maxLen int) string {
	checksum := md5.Sum(data)
	return base64.RawURLEncoding.EncodeToString(checksum[:maxLen])
}
