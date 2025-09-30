package minifier

import (
	"errors"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

type MockService struct {
	Config
	OriginalURLs map[string]string // key -- original URL, value -- minified ID
	MinifiedIDs  map[string]string // key -- minified ID, value -- original URL
}

// NewMockMinifier initializes MockMiinifier's fields used for testing:
//   - OriginalURLs
//   - MinifiedIDs
func NewMockMinifier() *MockService {
	return &MockService{
		OriginalURLs: make(map[string]string),
		MinifiedIDs:  make(map[string]string),
	}
}

func (s *MockService) MinifyURL(url string) (minifiedURL string, err error) {
	minifiedID, ok := s.OriginalURLs[url]
	if !ok {
		return "", errors.New("error: Unexpected original URL")
	}

	if minifiedID != encode([]byte(url), s.MaxLen) {
		return "", errors.New("error: Unexpected minify result")
	}

	return s.Config.BaseURL().String() + "/" + minifiedID, nil
}

func (s *MockService) UnMinifyURL(id string) (url string, err error) {
	origianlURL, ok := s.MinifiedIDs[id]
	if !ok {
		return "", storageErrors.NotFound
	}

	return origianlURL, nil
}
