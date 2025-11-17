// Package minifier is an implementation of [server.Service]
package minifier

import (
	"fmt"
	"net/url"
)

// Config contains fields and [flag.Value]s to set up the URL minifier
type Config struct {
	baseURL baseURL // domain parameter because the http server host address could be different
	MaxLen  int
}

// BaseURL returns a pointer to baseURL unexported type. Getter is used in case the structure of baseURL changes
func (c *Config) BaseURL() *baseURL { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.baseURL
}

type baseURL struct {
	scheme string
	host   string
	port   string
	Source string
}

func (b baseURL) String() string {
	return b.scheme + "://" + b.host + ":" + b.port
}

func (b *baseURL) Set(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return err
	}

	if url.Scheme == "" {
		return fmt.Errorf("%w: %s", errBaseURL, "empty scheme")
	}
	if url.Scheme != "http" {
		return fmt.Errorf("%w: %s", errBaseURL, "scheme MUST be 'http'")
	}
	b.scheme = url.Scheme

	if b.host = url.Hostname(); url.Hostname() == "" {
		return fmt.Errorf("%w: %s", errBaseURL, "empty host")
	}

	if b.port = url.Port(); url.Port() == "" {
		return fmt.Errorf("%w: %s", errBaseURL, "empty port")
	}

	return nil
}
