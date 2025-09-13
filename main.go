package main

import (
	"log"
	"net/http"
)

const defaultTCPPort string = "8080"

func minifyURL(res http.ResponseWriter, req *http.Request)   {}
func unminifyURL(res http.ResponseWriter, req *http.Request) {}

func init() {
	http.HandleFunc("POST /", minifyURL)
	http.HandleFunc("GET /{id}", unminifyURL)
}

func main() {
	srv := &http.Server{
		Addr: ":" + defaultTCPPort,
	}
	log.Printf("Minifier is listening on address: %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
