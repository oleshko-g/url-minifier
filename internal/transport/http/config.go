package http //revive:disable-line:var-naming

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// errParsingAddress indicates an error while parsing an address URL for an instance of http server
//
// [url-minifier]: https://github.com/oleshko-g/url-minifier
var errParsingAddress = errors.New("error parsing address")

// Config contains fields and [flag.Value]s to set up the [Server]
type Config struct {
	address       address
	canDecompress map[coding]struct{}
	canCompress   codings // MUST contain at least one element. [codingIdentity] MUST be the last element
	secretKey     secret
	secured       secured
	auditFile
	auditURL
}

// Address returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) Address() *address { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.address
}

// SecretAuthKey returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) SecretAuthKey() *secret { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.secretKey
}

// AuditFile returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) AuditFile() *auditFile { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.auditFile
}

// AuditURL returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) AuditURL() *auditURL { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.auditURL
}

// Secured returns a pointer to the [flag.Value] to set up the [Server]
func (c *Config) Secured() *secured { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.secured
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
		return fmt.Errorf("%w: %s", errParsingAddress, "empty scheme")
	}

	if a.port = url.Port(); url.Port() == "" {
		return fmt.Errorf("%w: %s", errParsingAddress, "empty port")
	}

	return nil
}

type secret string

// String is secret
func (sec secret) String() string {
	return ""
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

func (se *secured) String() string {
	if se != nil {
		return fmt.Sprint(*se)
	}

	return ""
}
