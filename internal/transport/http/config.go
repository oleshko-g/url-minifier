package http

import (
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	address       address
	canDecompress map[coding]struct{}
	canCompress   map[coding]struct{}
}

func (c *Config) Address() *address {
	return &c.address
}

type address struct {
	host   string
	port   string
	Source string
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

	if a.host = url.Hostname(); url.Hostname() == "" {
		return errors.New("error parsing address. empty host")
	}
	if a.port = url.Port(); url.Port() == "" {
		return errors.New("error parsing address. empty port")
	}

	return nil
}
