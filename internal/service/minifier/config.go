// Package minifier is an implementation of [server.Service]
package minifier

import (
	"fmt"
	"net/url"

	"github.com/oleshko-g/url-minifier/internal/config"
)

// NewConfig returns a new [Config] with the default values
func NewConfig() *Config {
	defaultBaseURL := "http://localhost:8080"
	return &Config{
		BaseURL: config.Option[*baseURL]{
			Name:        "b",
			Value:       new(baseURL),
			Description: fmt.Sprintf("Default: `%s`. Set the base URL for minified URLs", defaultBaseURL),
			Default:     defaultBaseURL,
		},
		MaxLen: 8,
	}
}

// Config contains fields and [flag.Value]s to set up the URL minifier
type Config struct {
	BaseURL config.Option[*baseURL] // domain parameter because the http server host address could be different
	MaxLen  int
}

type baseURL struct {
	scheme string
	host   string
	port   string
}

func (b *baseURL) String() string {
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
