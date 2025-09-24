package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"flag"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
)

var (
	minifiedURLs = make(map[string]string)
	cfg          = defaultConfig
)

func init() {
	flag.Var(&cfg.a, "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
	flag.Var(&cfg.b, "b", "Default: `https://localhost:8080`. Set the base URL for minified URLs")
}

func main() {
	flag.Parse()
	r := chi.NewRouter()
	r.Post("/", minifyURLHandler)
	r.Get("/{id}", unMinifyURLHandler)
	srv := &http.Server{
		Addr:    cfg.a.String(),
		Handler: r,
	}
	log.Printf("Minifier is listening on address: %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func minifyURLHandler(res http.ResponseWriter, req *http.Request) {
	err := validateContentType(req.Header, "text/plain")
	if err != nil {
		respondBadRequest(res, err)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
		return
	}

	_, err = url.Parse(string(data))
	if err != nil {
		respondBadRequest(res, err)
		return
	}

	minifiedURL := cfg.b.String() + "/" + encode(data)
	log.Printf("minifiedURL: %s", minifiedURL)
	minifiedURLs[encode(data)] = string(data)
	log.Printf("minifiedURLs:\n%+v\n", minifiedURLs)
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len(minifiedURL)))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(minifiedURL))
	log.Printf("%+v\n", res)
}

func validateContentType(h http.Header, value string) error {
	mediaType, _, errParseMediaType := mime.ParseMediaType(h.Get("Content-Type"))
	if errParseMediaType != nil {
		return errParseMediaType
	}

	if mediaType != value {
		return errors.New("error: Content-Type isn't " + value)
	}

	return nil
}

func encode(data []byte) string {
	checksum := md5.Sum(data)
	return base64.RawURLEncoding.EncodeToString(checksum[:cfg.maxLen])
}

func respondBadRequest(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	log.Printf("error: %s", err)
}

func unMinifyURLHandler(res http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		respondBadRequest(res, errors.New("empty id"))
		return
	}

	URL, ok := minifiedURLs[id]
	if !ok {
		respondBadRequest(res, errors.New("URL is not found"))
		return
	}
	res.Header().Add("Location", URL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusTemporaryRedirect)
	log.Printf("%#v\n", res)
}
