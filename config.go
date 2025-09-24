package main

import (
	"errors"
	"net/url"
	"strings"
)

var defaultConfig = config{
	a: address{
		host: "localhost",
		port: "8080",
	},
	b: baseURL{
		scheme: "http",
		address: address{
			host: "localhost",
			port: "8080",
		},
	},
	maxLen: 8,
}

type config struct {
	a      address
	b      baseURL
	maxLen int
}

type address struct {
	host string
	port string
}

func (a address) String() string {
	return a.host + ":" + a.port
}

func (a *address) Set(s string) error {
	if strings.HasPrefix(s, "localhost:") {
		s = "http://" + s
	}
	url, err := url.Parse(s)
	if err != nil {
		return err
	}
	if url.OmitHost {
		return errors.New("error parsing address. empty host")
	}
	if url.Port() == "" {
		return errors.New("error parsing address. empty port")
	}

	a.host = url.Hostname()
	a.port = url.Port()

	return nil
}

type baseURL struct {
	scheme string
	address
}

func (b baseURL) String() string {
	return b.scheme + "://" + b.address.String()
}

func (b *baseURL) Set(s string) error {
	err := b.address.Set(s)
	if err != nil {
		return err
	}
	url, err := url.Parse(s)
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		return errors.New("error parsing base URL. empty scheme")
	}
	if url.Scheme != "https" {
		return errors.New("error parsing base URL. scheme MUST be 'https'")
	}

	b.scheme = url.Scheme

	return nil
}
