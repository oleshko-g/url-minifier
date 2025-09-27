// import (
// 	"errors"
// 	"flag"
// 	"net/url"
// 	"strings"

// 	"github.com/oleshko-g/url-minifier/internal/minifier"
// )

// func init() {
// 	defaultConfig.baseURL.Set("http://localhost:8080")
// }

// var defaultConfig = config{
// 	a:       address{host: "localhost", port: "8080"},
// 	baseURL: &minifier.BaseURL{},
// 	maxLen:  8,
// }

// // type config struct {
// // 	a       address
// // 	baseURL flag.Getter
// // 	maxLen  int
// // }

// type address struct {
// 	host string
// 	port string
// }

// func (a address) String() string {
// 	return a.host + ":" + a.port
// }

// func (a *address) Set(s string) error {
// 	if strings.HasPrefix(s, "localhost:") {
// 		s = "http://" + s
// 	}
// 	url, err := url.Parse(s)
// 	if err != nil {
// 		return err
// 	}
// 	if url.OmitHost {
// 		return errors.New("error parsing address. empty host")
// 	}
// 	if url.Port() == "" {
// 		return errors.New("error parsing address. empty port")
// 	}

// 	a.host = url.Hostname()
// 	a.port = url.Port()

// 	return nil
// }

// // type baseURL struct {
// // 	scheme string
// // 	address
// // }

// // func (b baseURL) String() string {
// // 	return b.scheme + "://" + b.address.String()
// // }

// // func (b *baseURL) Set(s string) error {
// // 	err := b.address.Set(s)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	url, err := url.Parse(s)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	if url.Scheme == "" {
// // 		return errors.New("error parsing base URL. empty scheme")
// // 	}
// // 	if url.Scheme != "http" {
// // 		return errors.New("error parsing base URL. scheme MUST be 'http'")
// // 	}

// // 	b.scheme = url.Scheme

// // 	return nil
// // }
package main
