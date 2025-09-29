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
	minifier.Storage
	*minifier.Service
	*http.Server
}

func (a *app) setup() error {
	a.Storage = memory.NewStrRecords()
	log.Print("storage set")

	a.Service = minifier.New(a.Storage)
	a.Service.Config.MaxLen = 8
	a.Service.Config.BaseURL().Set("http://localhost:8080")
	flag.Var(a.Service.Config.BaseURL(), "b", "Default: `http://localhost:8080`. Set the base URL for minified URLs")
	log.Printf("base URL is set: %s", a.Service.Config.BaseURL().String())

	a.Server = http.NewServer(a.Service)
	a.Server.Config.Address().Set("localhost:8080")
	flag.Var(a.Server.Config.Address(), "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")

	flag.Parse()

	return nil
}
