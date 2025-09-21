package main

import (
	"errors"
	"flag"
	"net/url"
)

var defaultConfig = minifierConfig{
	a: minifierAddress{
		host: "localhost",
		port: "8080",
	},
	b: minifierBaseURL{
		scheme: "https://",
		minifierAddress: minifierAddress{
			host: "localhost",
			port: "8080",
		},
	},
}

func newMinifierConfig() (minifierConfig, error) {
	var cfg minifierConfig
	flag.CommandLine.Var(&cfg.a, "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
	flag.CommandLine.Var(&cfg.b, "b", "Default: `https://localhost:8080`. Set the base URL for minified URLs")
	flag.Parse()
	return cfg, nil
}

type minifierConfig struct {
	a minifierAddress
	b minifierBaseURL
}

type minifierAddress struct {
	host string
	port string
}

func (a minifierAddress) String() string {
	return a.host + ":" + a.port
}

func (a minifierAddress) Set(s string) error {
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

type minifierBaseURL struct {
	scheme string
	minifierAddress
}

func (b minifierBaseURL) String() string {
	return b.scheme + b.minifierAddress.String()
}

func (b minifierBaseURL) Set(s string) error {
	err := b.minifierAddress.Set(s)
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
	if url.Scheme != "https://" {
		return errors.New("error parsing base URL. scheme MUST be 'https://'")
	}

	b.scheme = url.Scheme

	return nil
}
