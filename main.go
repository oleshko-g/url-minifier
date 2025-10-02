package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
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

func (a *app) setup() (err error) {
	godotenv.Load(".env")

	a.Storager = memory.NewStrRecords()
	log.Print("storage set")

	a.Service = minifier.New(a.Storager)

	// Set the default config values
	a.Service.Config.MaxLen = 8
	a.Service.Config.BaseURL().Set("http://localhost:8080")

	a.Server = http.NewServer(a.Service)
	a.Server.Config.Address().Set("localhost:8080")

	// If an env var is present then it overrides the default value or the flag value
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		// sets err func (a *app) setup()
		if err = a.Service.Config.BaseURL().Set(baseURL); err != nil {
			return
		}
		a.Service.Config.BaseURL().Source = "ENV"
	}
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		if err = a.Server.Config.Address().Set(serverAddress); err != nil {
			return
		}
		a.Server.Config.Address().Source = "ENV"
	}

	// sets flags
	flag.Var(a.Service.Config.BaseURL(), "b", "Default: `http://localhost:8080`. Set the base URL for minified URLs")
	flag.Var(a.Server.Config.Address(), "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
	flag.Parse()

	log.Printf("Base URL is set to `%s`", a.Service.Config.BaseURL().String())
	log.Printf("Server Address is set to `%s`", a.Server.Config.Address().String())

	return nil
}
