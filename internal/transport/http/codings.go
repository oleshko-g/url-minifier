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
func parseAcceptEncoding(h http.Header) (parsedAcceptCodings map[coding]qualityValue, err error) {
	acceptEncodingValues := h.Values("Accept-Encoding")
	if acceptEncodingValues == nil {
		return parsedAcceptCodings, nil // return the nil map with no error, "no Accept-Encoding"
	}

	parsedAcceptCodings = make(map[coding]qualityValue)

	for _, v := range acceptEncodingValues {
		pe, err := parseCoding(v)
		if err != nil {
			// "Accept-Encoding" with an only empty string value means implicit "identity;q=1.0". Break the loop with no error
			if errors.Is(err, errEmptyCoding) && len(acceptEncodingValues) == 1 {
				break
			}
			return nil, err
		}

		parsedAcceptCodings[pe.coding] = pe.qualityValue
	}

	// if [codingIdentity] is not specified explicitly then set it with the default [qualityValue]
	if _, ok := parsedAcceptCodings[codingIdentity]; !ok {
		parsedAcceptCodings[codingIdentity] = 1.0
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

	if !coding(s).valid() {
		return parsedCoding{},
			fmt.Errorf("%s is invalid coding", s)
	}
	pe.coding = coding(s)

	if !found { // if ";q=" is not set, then qualityValue is 1.000
		pe.qualityValue = 1.000
		return pe, nil
	}

	// parse quality value
	if strings.ToLower(after) == "nan" ||
		strings.ToLower(after) == "inf" ||
		strings.ToLower(after) == "infinity" {
		defer func() { _ = pe }()
		return parsedCoding{},
			fmt.Errorf("quality value %s of %s coding doesn't satisfy '0.0 <= quality value <= 1.0'", after, pe.coding)
	}
	parsedFloat, err := strconv.ParseFloat(after, 32)
	if err != nil {
		defer func() { _, _ = pe, parsedFloat }()
		return parsedCoding{},
			errors.Join(errors.New("failed to parse quality value '%s' of %s coding"), err)
	}
	if parsedFloat < 0.0 || parsedFloat < 1.0 {
		defer func() { _, _ = pe, parsedFloat }()
		return parsedCoding{},
			fmt.Errorf("quality value %s of %s coding doesn't satisfy '0.0 <= quality value <= 1.0'", after, pe.coding)
	}

	// TODO: add the "up to 3 decimal digits" check

	pe.qualityValue = qualityValue(parsedFloat)
	return pe, nil
}

var errEmptyCoding = errors.New("empty coding value")

// parsedCoding represents a valid HTTP coding value used in Content-Encoding or Accept-Encoding HTTP header fields
type parsedCoding struct {
	coding
	qualityValue // default is 1.0
}

type (
	qualityValue float32 // Quality values, or q-values and q-factors, are used to describe the order of priority of values in a comma-separated list. It is a special syntax allowed in some HTTP headers and in HTML.
	coding       string  // coding is a string which might be a valid HTTP Content coding listed in [validCodings]
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
	codingWildcard    coding = "*"            // wildcard which means "any other coding besides others specified"
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
