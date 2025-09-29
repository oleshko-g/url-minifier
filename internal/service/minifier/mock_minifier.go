package minifier

import (
	"errors"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

type MockService struct {
	MaxLen       int
	originalURLs map[string]string // key -- original URL, value -- minified ID
	minifiedIDs  map[string]string // key -- minified ID, value -- original URL
}

func (s *MockService) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedID, ok := s.originalURLs[url]
	if !ok {
		return "", errors.New("error: Unexpected original URL")
	}

	if minifiedID != encode([]byte(url), s.MaxLen) {
		return "", errors.New("error: Unexpected minify result")
	}

	return "http://localhost:8080/" + minifiedID, nil
}

func (s *MockService) UnMinifyURL(id string) (url string, err error) {
	origianlURL, ok := s.minifiedIDs[id]
	if !ok {
		return "", storageErrors.NotFound
	}

	return origianlURL, nil
}
