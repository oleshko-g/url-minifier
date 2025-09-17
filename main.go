package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
)

const defaultTCPPort string = "8080"

func init() {
	http.HandleFunc("POST /", minifyURLHandler)
	http.HandleFunc("GET /{id}", unMinifyURLHandler)
}

var minifiedURLs = make(map[string]string)

func validateContenType(h http.Header, value string) error {
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
	return base64.RawURLEncoding.EncodeToString(checksum[:8])
}

func respondBadRequest(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	log.Printf("error: %s", err)
}

func minifyURLHandler(res http.ResponseWriter, req *http.Request) {
	err := validateContenType(req.Header, "text/plain")
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

	minifiedURL := "http://" + req.Host + "/" + encode(data)
	log.Printf("minifiedURL: %s", minifiedURL)
	minifiedURLs[encode(data)] = string(data)
	log.Printf("minifiedURLs:\n%+v\n", minifiedURLs)
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len(minifiedURL)))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(minifiedURL))
	log.Printf("%+v\n", res)
}

func main() {
	srv := &http.Server{
		Addr: ":" + defaultTCPPort,
	}
	log.Printf("Minifier is listening on address: %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
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
