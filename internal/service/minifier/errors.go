package minifier

import "errors"

// errBaseURL indicates an error while parsing a base URL for an instance of [url-minifier]
//
// [url-minifier]: https://github.com/oleshko-g/url-minifier
var errBaseURL = errors.New("error parsing base URL")
