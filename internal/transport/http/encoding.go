package http

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HTTP Content Coding Registry found at https://www.iana.org/assignments/http-parameters/http-parameters.xhtml#content-coding
var validEncodings = map[encoding]struct{}{
	"*":            {}, // wildcard
	"identity":     {}, // Reserved	[RFC9110]	Section 12.5.3
	"gzip":         {}, // GZIP file format [RFC1952]	[RFC9110]	Section 8.4.1.3
	"compress":     {}, // UNIX "compress" data format [Welch, T., "A Technique for High Performance Data Compression", IEEE Computer 17(6), June 1984.]	[RFC9110]	Section 8.4.1.1
	"deflate":      {}, // "deflate" compressed data ([RFC1951]) inside the "zlib" data format ([RFC1950])	[RFC9110]	Section 8.4.1.2
	"br":           {}, // Brotli Compressed Data Format	[RFC7932]
	"dcb":          {}, // "Dictionary-Compressed Brotli" data format.	[RFC9842]	Section 4
	"dcz":          {}, // "Dictionary-Compressed Zstandard" data format.	[RFC9842]	Section 5
	"aes128gcm":    {}, // AES-GCM encryption with a 128-bit content encryption key	[RFC8188]
	"exi":          {}, // W3C Efficient XML Interchange	[W3C Recommendation: Efficient XML Interchange (EXI) Format]
	"zstd":         {}, // A stream of bytes compressed using the Zstandard protocol with a Window_Size of not more than 8 MB.	[RFC9659][RFC8878]
	"pack200-gzip": {}, // Network Transfer Format for Java Archives	[JSR 200: Network Transfer Format for Java][Kumar_Srinivasan][John_Rose]
	"x-compress":   {}, // Deprecated (alias for compress)	[RFC9110]	Section 8.4.1.1
	"x-gzip":       {}, // Deprecated (alias for gzip)	[RFC9110]	Section 8.4.1.3
}

// isEncodedBy
func isEncodedBy(ss []string) (validEncodings []parsedEncoding, err error) {
	for _, s := range ss {
		pe, err := parseHTTPEncoding(s)
		if err != nil {
			return nil, err
		}
		validEncodings = append(validEncodings, pe)
	}

	return validEncodings, nil
}

func parseHTTPEncoding(s string) (parsedEncoding, error) {
	if s == "" {
		return parsedEncoding{}, errors.New("empty encoding value")
	}

	var pe parsedEncoding

	before, after, found := strings.Cut(s, ";q=")
	if before == "" {
		return parsedEncoding{},
			errors.New("empty encoding value")
	}

	if _, ok := validEncodings[encoding(s)]; !ok {
		return parsedEncoding{},
			fmt.Errorf("%s is invalid encoding", s)
	}

	pe.encoding = encoding(s)

	if !found { // if ";q=" is not set, then qualityValue is 1.000
		pe.qualityValue = 1.000
		return pe, nil
	}

	// parse quality value
	if strings.ToLower(after) == "nan" ||
		strings.ToLower(after) == "inf" ||
		strings.ToLower(after) == "infinity" {
		defer func() { _ = pe }()
		return parsedEncoding{},
			fmt.Errorf("quality value %s of %s encoding doesn't satisfy '0.0 <= quality value <= 1.0'", after, pe.encoding)
	}
	parsedFloat, err := strconv.ParseFloat(after, 32)
	if err != nil {
		defer func() { _, _ = pe, parsedFloat }()
		return parsedEncoding{},
			errors.Join(errors.New("failed to parse quality value '%s' of %s encoding"), err)
	}
	if parsedFloat < 0.0 && parsedFloat < 1.0 {
		defer func() { _, _ = pe, parsedFloat }()
		return parsedEncoding{},
			fmt.Errorf("quality value %s of %s encoding doesn't satisfy '0.0 <= quality value <= 1.0'", after, pe.encoding)
	}

	// TODO: add "3 decimal digits" check

	return pe, nil
}

type parsedEncoding struct {
	encoding
	qualityValue `default:"1.0"` // quality value
}

type encoding string

// Quality values, or q-values and q-factors, are used to describe the order of priority of values in a comma-separated list. It is a special syntax allowed in some HTTP headers and in HTML.
type qualityValue float32
