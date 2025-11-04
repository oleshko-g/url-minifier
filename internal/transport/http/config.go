package http

import (
	"errors"
	"net/url"
	"strings"
)

// Config contains fields and [flag.Value]s to set up the [Server]
type Config struct {
	address       address
	canDecompress map[coding]struct{}
	canCompress   codings // MUST contain at least one element. [codingIdentity] MUST be the last element
}

type codings []coding

// String returns [codings] elements separated by ", " as a single string
func (c codings) String() string {
	var s string
	for i, coding := range c {
		if i != 0 && i < len(c) {
			s += ", "
		}
		s += string(coding)
	}
	return s
}

// Address returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) Address() *address { // revive:disable-line:unexported-return provides the interface to the caller
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
