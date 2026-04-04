package http //revive:disable-line:var-naming

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
			Name:        "a",
			Value:       new(address),
			Default:     "localhost:8080",
			Description: "Sets the network address and the port for the minifier",
		},
		AuditFile: config.Option[*auditFile]{
			Name:        "audit-file",
			Value:       new(auditFile),
			Description: "Sets the file to write audit logs to",
		},
		AuditURL: config.Option[*auditURL]{
			Name:        "audit-url",
			Value:       new(auditURL),
			Description: "Sets the URL to write audit logs to",
		},
		Secured: config.Option[*secured]{
			Name:        "s",
			Value:       new(secured),
			Description: "Sets the \"secured\" flag. If set the minifier HTTP server listens using TLS protocol",
		},
		SecretKey: config.Option[*secret]{
			Value: new(secret),
		},
	}

	cfg.Address.Set(cfg.Address.Default)
	cfg.Address.Source = config.SourceDefault

	cfg.SecretKey.Set(cfg.SecretKey.Default)
	cfg.SecretKey.Source = config.SourceDefault

	cfg.Secured.Set(cfg.Secured.Default)
	cfg.Secured.Source = config.SourceDefault

	cfg.AuditFile.Set(cfg.AuditFile.Default)
	cfg.AuditFile.Source = config.SourceDefault

	cfg.AuditURL.Set(cfg.AuditURL.Default)
	cfg.AuditURL.Source = config.SourceDefault

	return &cfg
}

// Config contains fields and [flag.Value]s to set up the [Server]
type Config struct {
	Address       config.Option[*address]
	SecretKey     config.Option[*secret]
	Secured       config.Option[*secured]
	AuditFile     config.Option[*auditFile]
	AuditURL      config.Option[*auditURL]
	TrustedSubnet config.Option[*trustedIPsubnet]
	canCompress   codings // MUST contain at least one element. [codingIdentity] MUST be the last element
	canDecompress map[coding]struct{}
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

type secured bool

func (se *secured) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	*se = secured(v)
	return nil
}

func (se secured) String() string {
	// current se value assignable to bool
	return strconv.FormatBool(bool(se))
}
