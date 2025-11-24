// Package http is an internal [http.Server] implementation
package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// parseContentEncoding parses "Content-Encoding" HTTP header.
//   - If "Content-Encoding" is absent it return a nil [parsedContentEncodings] slice and a nil error
//   - Otherwise it parses "Content-Encoding" existing values and returns a populated [parsedContentCodings]---ordered the same as "Content-Encoding" values---slice  or a parsing error
func parseContentEncoding(h http.Header) (parsedContentCodings []parsedCoding, err error) {
	contentEncodingValues := h.Values("Content-Encoding")
	if contentEncodingValues == nil {
		return parsedContentCodings, nil // no "Content-Encodings, return empty
	}

	for _, v := range contentEncodingValues {
		pe, err := parseCoding(v)
		if err != nil {
			return nil, err
		}
		parsedContentCodings = append(parsedContentCodings, pe)
	}

	return parsedContentCodings, nil
}

// parseAcceptEncoding parses "Accept-Encoding" HTTP header
//   - If "Accept-Encoding" is absent it returns a nil [parsedAcceptCodings] map and a nil error
//   - Otherwise it parses "Accept-Encoding" existing values and returns a populated [parsedAcceptCodings] map or a parsing error
//   - If [codingIdentity] is not specified explicitly it sets it in the [parsedAcceptCodings] map as if "identity;q=1.0" was parsed
//
// TODO: add tests
func parseAcceptEncoding(h http.Header) (parsedAcceptCodings map[coding]qualityValue, err error) {
	acceptEncodingValues := h.Values("Accept-Encoding")
	if acceptEncodingValues == nil {
		return parsedAcceptCodings, nil // return the nil map with no error, "no Accept-Encoding"
	}

	ss := strings.Split(acceptEncodingValues[0], ",")

	parsedAcceptCodings = make(map[coding]qualityValue)

	for _, s := range ss {
		pe, err := parseCoding(strings.TrimSpace(s))
		if err != nil {
			// "Accept-Encoding" with an only empty string value means implicit "identity;q=1.0". Set it and break the loop with no error
			if errors.Is(err, errEmptyCoding) && len(acceptEncodingValues) == 1 {
				parsedAcceptCodings[codingIdentity] = 1.0
				break
			}
			return nil, err
		}

		parsedAcceptCodings[pe.coding] = pe.qualityValue
	}

	return parsedAcceptCodings, nil
}

func parseCoding(s string) (parsedCoding, error) {
	if s == "" {
		return parsedCoding{}, errEmptyCoding
	}

	var pe parsedCoding

	before, after, found := strings.Cut(s, ";q=")
	if before == "" {
		return parsedCoding{},
			errEmptyCoding
	}

	if !coding(before).valid() {
		return parsedCoding{},
			fmt.Errorf("%s is invalid coding", s)
	}
	pe.coding = coding(before)

	if !found { // if ";q=" is not set, then qualityValue is 1.000
		pe.qualityValue = 1.000
		return pe, nil
	}

	// parse quality value
	parsedFloat, err := parseQualityValue(after)
	if err != nil {
		defer func() { _ = pe }()
		switch err {
		case errQualityValueOutOfRange:
			err = fmt.Errorf("%w. quality value %s of %s coding doesn't satisfy '0.0 <= quality value <= 1.0'", err, after, pe.coding)
		case errQualityValueTooLong:
			err = fmt.Errorf("%w. quality value %s of %s coding has more then 3 digits after the point", err, after, pe.coding)
		case errQualityValueParseFloat:
			err = fmt.Errorf("%w. failed to parse quality value '%s' of %s coding", err, after, pe.coding)
		}
		return parsedCoding{}, err
	}
	pe.qualityValue = qualityValue(parsedFloat)
	return pe, nil
}

var errEmptyCoding = errors.New("empty coding value")

// parseQualityValue validate quality value and can return:
//   - [err]
func parseQualityValue(s string) (float64, error) {
	// parse quality value
	if strings.ToLower(s) == "nan" ||
		strings.ToLower(s) == "inf" ||
		strings.ToLower(s) == "infinity" {
		return 0.0, errQualityValueOutOfRange
	}

	if len(s) > 5 {
		return 0.0, errQualityValueTooLong
	}

	parsedFloat, err := strconv.ParseFloat(s, 32)
	if err != nil {
		defer func() { _ = parsedFloat }()
		return 0.0, errQualityValueParseFloat
	}
	if parsedFloat < 0.0 && parsedFloat < 1.0 {
		defer func() { _ = parsedFloat }()
		return 0.0, errQualityValueOutOfRange
	}
	return parsedFloat, nil
}

var (
	errQualityValueOutOfRange = errors.New("quality value is out of the allowed range")
	errQualityValueTooLong    = errors.New("quality value is too long")
	errQualityValueParseFloat = errors.New("quality value couldn't be parsed to float ")
)

// parsedCoding represents a valid HTTP coding value used in Content-Encoding or Accept-Encoding HTTP header fields
type parsedCoding struct {
	coding
	qualityValue
}

type (
	// qualityValue is relative "weight" of a content coding in a comma-separated list. Syntax: ";q=[qualityValue]". Default: "1.0".
	// [Quality values]: https://httpwg.org/specs/rfc9110.html#quality.values
	qualityValue float32
	// coding is a string which might be a valid HTTP Content coding listed in [validCodings].
	coding string
)

func (c coding) valid() bool {
	_, ok := validCodings[c]
	return ok
}

// HTTP Content Coding Registry found at https://www.iana.org/assignments/http-parameters/http-parameters.xhtml#content-coding
var validCodings = map[coding]struct{}{
	codingWildcard:    {},
	codingIdentity:    {},
	codingGZIP:        {},
	codingCompress:    {},
	codingDeflate:     {},
	codingBrotli:      {},
	codingDCB:         {},
	codingDCZ:         {},
	codingAES128GCM:   {},
	codingEXI:         {},
	codingZstd:        {},
	codingPack200GZIP: {},
	codingXCompress:   {},
	codingXGZIP:       {},
}

const (
	codingWildcard    coding = "*"            // The wildcard. Which means "any other coding besides the specified"
	codingIdentity    coding = "identity"     // Reserved	[RFC9110]	Section 12.5.3
	codingGZIP        coding = "gzip"         // GZIP file format [RFC1952]	[RFC9110]	Section 8.4.1.3
	codingCompress    coding = "compress"     // UNIX "compress" data format [Welch, T., "A Technique for High Performance Data Compression", IEEE Computer 17(6), June 1984.]	[RFC9110]	Section 8.4.1.1
	codingDeflate     coding = "deflate"      // "deflate" compressed data ([RFC1951]) inside the "zlib" data format ([RFC1950])	[RFC9110]	Section 8.4.1.2
	codingBrotli      coding = "br"           // Brotli Compressed Data Format	[RFC7932]
	codingDCB         coding = "dcb"          // "Dictionary-Compressed Brotli" data format.	[RFC9842]	Section 4
	codingDCZ         coding = "dcz"          // "Dictionary-Compressed Zstandard" data format.	[RFC9842]	Section 5
	codingAES128GCM   coding = "aes128gcm"    // AES-GCM encryption with a 128-bit content encryption key	[RFC8188]
	codingEXI         coding = "exi"          // W3C Efficient XML Interchange	[W3C Recommendation: Efficient XML Interchange (EXI) Format]
	codingZstd        coding = "zstd"         // A stream of bytes compressed using the Zstandard protocol with a Window_Size of not more than 8 MB.	[RFC9659][RFC8878]
	codingPack200GZIP coding = "pack200-gzip" // Network Transfer Format for Java Archives	[JSR 200: Network Transfer Format for Java][Kumar_Srinivasan][John_Rose]
	codingXCompress   coding = "x-compress"   // Deprecated: (alias for compress)	[RFC9110]	Section 8.4.1.1
	codingXGZIP       coding = "x-gzip"       // Deprecated: (alias for gzip)	[RFC9110]	Section 8.4.1.3
)
