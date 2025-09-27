package minifier

import (
	"errors"
	"net/url"
)

type Config struct {
	baseURL baseURL
	MaxLen  int
}

func (c *Config) BaseURL() *baseURL {
	return &c.baseURL
}

type baseURL struct {
	scheme string
	host   string
	port   string
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
		return errors.New("error parsing base URL. empty scheme")
	}
	if url.Scheme != "http" {
		return errors.New("error parsing base URL. scheme MUST be 'http'")
	}
	b.scheme = url.Scheme

	if b.host = url.Hostname(); url.Hostname() == "" {
		return errors.New("error parsing address. empty host")
	}

	if b.port = url.Port(); url.Port() == "" {
		return errors.New("error parsing address. empty port")
	}

	return nil
}
