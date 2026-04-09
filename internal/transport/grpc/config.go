package grpc //revive:disable-line:var-naming

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/oleshko-g/url-minifier/internal/config"
)

// errParsingAddress indicates an error while parsing an address URL for an instance of http server
//
// [url-minifier]: https://github.com/oleshko-g/url-minifier
var errParsingAddress = errors.New("error parsing address")

// NewConfig returns a default HTTP server [Config].
func NewConfig() *Config {
	cfg := Config{
		Address: config.Option[*address]{
			Name:        "grpc_address",
			EnVarName:   "GRPC_ADDRESS",
			Value:       new(address),
			Default:     "localhost:8081",
			Description: "Sets the network address and the port for the minifier",
		},
		SecretKey: config.Option[*secret]{
			Value: new(secret),
			Description: "Secret key is used to sign user auth tokens.",
		},
	}

	cfg.Address.Set(cfg.Address.Default)
	cfg.Address.Source = config.SourceDefault

	cfg.SecretKey.Set(cfg.SecretKey.Default)
	cfg.SecretKey.Source = config.SourceDefault
	return &cfg
}

// Config contains fields and [flag.Value]s to set up the [Server]
type Config struct {
	Address   config.Option[*address]
	SecretKey config.Option[*secret]
}

type address struct {
	host string
	Port string
}

func (a *address) String() string {
	return a.host + ":" + a.Port
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
		return fmt.Errorf("%w: %s", errParsingAddress, "empty host")
	}

	if a.Port = url.Port(); url.Port() == "" {
		return fmt.Errorf("%w: %s", errParsingAddress, "empty port")
	}

	return nil
}

type secret string

// String is secret
func (sec secret) String() string {
	return string(sec)
}

// Set sets the secret key of the minifier
func (sec *secret) Set(s string) error {
	*sec = secret(s)
	return nil
}
