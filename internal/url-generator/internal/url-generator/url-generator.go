package urlgenerator

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"io"
	"math/rand"
	"strings"
	"sync"
)

const defaultURLPartLen = 2

func NewURLGenerator(w io.WriteCloser, urlPartLen int) *urlGenerator {
	return &urlGenerator{
		URLPartLen: defaultURLPartLen,
		rBuf:       bytes.NewBuffer(make([]byte, defaultURLPartLen)),
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
	ug.builder.WriteString(ug.randromAuthority())
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

	if genrateQuery := flipCoin(); genrateQuery {
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

func (ug *urlGenerator) randromAuthority() string {
	switch rand.Intn(2) {
	case 0:
		return "http"
	case 1:
		return "https"
	}
	return ""
}

func (ug *urlGenerator) randomURLEncodedString(len int) string {
	for ug.rBuf.Len() < len {
		ug.rBuf.WriteByte(0) // inittialize with zeroes to fill up by random bytes
	}

	cryptoRand.Read(ug.rBuf.Bytes()) //revive:disable-line Fills up the buffer and never returns an error
	defer ug.rBuf.Reset()

	return base64.RawURLEncoding.EncodeToString(ug.rBuf.Bytes()[:len])
}

func flipCoin() bool {
	r := rand.Int()
	return r%2 == 0
}
