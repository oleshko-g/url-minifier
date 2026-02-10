package http //revive:disable-line:var-naming

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
)

// errParsingAdress indicates an error while parsing an address URL for an instance of http server
//
// [url-minifier]: https://github.com/oleshko-g/url-minifier
var errParsingAdress = errors.New("error parsing address")

// Config contains fields and [flag.Value]s to set up the [Server]
type Config struct {
	address       address
	canDecompress map[coding]struct{}
	canCompress   codings // MUST contain at least one element. [codingIdentity] MUST be the last element
	secretKey     secret
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
		return fmt.Errorf("%w: %s", errParsingAdress, "empty scheme")
	}

	if a.port = url.Port(); url.Port() == "" {
		return fmt.Errorf("%w: %s", errParsingAdress, "empty port")
	}

	return nil
}

type secret string

// Steing is secret
func (sec secret) String() string {
	return ""
}

// Set sets the secret key of the minifier
func (sec *secret) Set(s string) error {
	*sec = secret(s)
	return nil
}

type auditFile struct {
	fp      *os.File
	mu      sync.Mutex
	enabled bool
	Source  string
}

// Set oprn or creates the audit file or returns an error
func (a *auditFile) Set(s string) error {
	// Write Read _, Read _ _, Read _ _
	const filePerm os.FileMode = 0o644

	fp, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		return err
	}

	a.fp = fp
	a.enabled = true
	return nil
}

// String returns the name of the audit file
func (a *auditFile) String() string {
	if a.fp == nil {
		return ""
	}

	return a.fp.Name()
}

type auditURL struct {
	url     *url.URL
	enabled bool
	Source  string
}

// Set parses s into a [url.URL] and sets it as the value of audit URL
func (a *auditURL) Set(s string) error {
	parsedURL, err := url.Parse(s)
	if err != nil {
		return err
	}

	a.url = parsedURL
	a.enabled = true
	return nil
}

// String return the opeque respresentation of an audit URL
func (a *auditURL) String() string {
	if a.url == nil {
		return ""
	}

	return a.url.Opaque
}
