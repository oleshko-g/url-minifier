package minifier

import "net/url"

type (
	URL struct {
		OriginalURL minifierURL
		MinifiedURL minifierURL
	}
	minifierURL interface {
		Parse(string) (*url.URL, error)
		String() string
	}
)
