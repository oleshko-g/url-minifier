// Package urlgenerator is a utility package that implements random URL generator
package urlgenerator

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"strings"
	"sync"
)

const defaultURLPartLen = 2

// NewURLGenerator returns return random URL generator with the set max URL part length to [Generate] repeatedly
func NewURLGenerator(urlPartLen int) *urlGenerator { // revive:disable-line:unexported-return return a pointer to repeatedly call [Generate]
	if urlPartLen <= 0 {
		urlPartLen = defaultURLPartLen
	}
	return &urlGenerator{
		URLPartLen: urlPartLen,
		rBuf:       bytes.NewBuffer(make([]byte, urlPartLen)),
	}
}

type urlGenerator struct {
	URLPartLen int

	builder strings.Builder
	rBuf    *bytes.Buffer

	mu sync.RWMutex
}

// Generate returns a string which is a randomly generated URL
func (ug *urlGenerator) Generate() string {
	ug.mu.Lock()
	defer ug.mu.Unlock()
	defer ug.builder.Reset()

	// generate Authority
	ug.builder.WriteString(ug.randomAuthority())
	ug.builder.WriteString("://")

	// generate Host
	ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
	ug.builder.WriteRune('.')
	ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))

	//generate a PathSegment?
	for flipCoin() {
		ug.builder.WriteRune('/')
		ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
	}

	if generateQuery := flipCoin(); generateQuery {
		// generate the first Query
		ug.builder.WriteRune('?')
		//key
		ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
		ug.builder.WriteRune('=')
		//value
		ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))

		// generate another Query?
		for flipCoin() {
			ug.builder.WriteRune('&')
			//key
			ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
			ug.builder.WriteRune('=')
			//value
			ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
		}
	}

	// generate Fragment?
	if flipCoin() {
		ug.builder.WriteRune('#')
		ug.builder.WriteString(ug.randomURLEncodedString(ug.URLPartLen))
	}

	return ug.builder.String()
}

func (ug *urlGenerator) randomAuthority() string {
	switch rand.Intn(2) {
	case 0:
		return "http"
	case 1:
		return "https"
	}
	return ""
}

func (ug *urlGenerator) randomURLEncodedString(length int) string {
	for ug.rBuf.Len() < length {
		ug.rBuf.WriteByte(0) // initialize with zeroes to fill up by random bytes
	}

	cryptoRand.Read(ug.rBuf.Bytes()) //revive:disable-line Fills up the buffer and never returns an error
	defer ug.rBuf.Reset()

	return base64.RawURLEncoding.EncodeToString(ug.rBuf.Bytes()[:length])
}

func flipCoin() bool {
	r := rand.Int()
	return r%2 == 0
}
