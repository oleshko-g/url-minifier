package http

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func parseContentCodings(ss []string) (validEncodings []parsedCoding, err error) {
	for _, s := range ss {
		pe, err := parseContentCoding(s)
		if err != nil {
			return nil, err
		}
		validEncodings = append(validEncodings, pe)
	}

	return validEncodings, nil
}

func parseContentCoding(s string) (parsedCoding, error) {
	if s == "" {
		return parsedCoding{}, errors.New("empty coding value")
	}

	var pe parsedCoding

	before, after, found := strings.Cut(s, ";q=")
	if before == "" {
		return parsedCoding{},
			errors.New("empty coding value")
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
	if parsedFloat < 0.0 && parsedFloat < 1.0 {
		defer func() { _, _ = pe, parsedFloat }()
		return parsedCoding{},
			fmt.Errorf("quality value %s of %s coding doesn't satisfy '0.0 <= quality value <= 1.0'", after, pe.coding)
	}

	// TODO: add the "up to 3 decimal digits" check

	pe.qualityValue = qualityValue(parsedFloat)
	return pe, nil
}

// parsedCoding represents a valid HTTP coding value used in Content-Encoding or Accept-Encoding HTTP header fields
type parsedCoding struct {
	coding
	qualityValue // default is 1.0
}

type (
	qualityValue float32 // Quality values, or q-values and q-factors, are used to describe the order of priority of values in a comma-separated list. It is a special syntax allowed in some HTTP headers and in HTML.
	coding       string  // coding is the alias a string which is possibly an HTTP Content coding listed in [validCodings]
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
