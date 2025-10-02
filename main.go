package main

import (
	"flag"
	"log"

	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage/memory"
	"github.com/oleshko-g/url-minifier/internal/transport/http"
)

var a app

func main() {
	if err := a.setup(); err != nil {
		log.Fatal(err)
	}
	log.Fatal(a.Server.ListenAndServe())
}

type app struct {
	minifier.Storager
	*minifier.Service
	*http.Server
}

func (a *app) setup() error {
	a.Storager = memory.NewStrRecords()
	log.Print("storage set")

	a.Service = minifier.New(a.Storager)

	// set the default config values
	a.Service.Config.MaxLen = 8
	a.Service.Config.BaseURL().Set("http://localhost:8080")

	a.Server = http.NewServer(a.Service)
	a.Server.Config.Address().Set("localhost:8080")

	// set up flags
	flag.Var(a.Service.Config.BaseURL(), "b", "Default: `http://localhost:8080`. Set the base URL for minified URLs")
	flag.Var(a.Server.Config.Address(), "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")

	// if an env var is empty then the Parse sets a config
	// if a flag is empty then the config keeps the default value
	flag.Parse()

	log.Printf("base URL is set: %s", a.Service.Config.BaseURL().String())
	log.Printf("server address is set: %s", a.Server.Config.Address().String())

	return nil
}
