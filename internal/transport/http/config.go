package http

import (
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	address Address
}

func (c *Config) Address() *Address {
	return &c.address
}

type Address struct {
	host string
	port string
}

func (a Address) String() string {
	return a.host + ":" + a.port
}

func (a *Address) Set(s string) error {
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
