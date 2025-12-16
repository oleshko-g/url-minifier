package minifier

import "net/url"

type (
	// URL is the structure intended to be returned by minifier Public methods
	URL struct {
		OriginalURL minifierURL
		MinifiedURL minifierURL
		Deleted     bool
	}
	minifierURL interface {
		Parse(string) (*url.URL, error)
		String() string
	}
)
