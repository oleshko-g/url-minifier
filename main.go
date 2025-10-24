package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage/file"
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
	fileConfig     file.Config
	minifierConfig minifier.Config
	httpConfig     http.Config
	minifier.Storager
	*minifier.Service
	*http.Server
}

func (a *app) setup() (err error) {
	// Set the default config values
	a.fileConfig.Path().Set(file.DefaultPath)
	a.minifierConfig.MaxLen = 8
	a.minifierConfig.BaseURL().Set("http://localhost:8080")
	a.httpConfig.Address().Set("localhost:8080")

	// If an env var is present then it overrides the default value or the flag value
	godotenv.Load(".env")
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		// sets err func (a *app) setup()
		if err = a.fileConfig.Path().Set(filePath); err != nil {
			return
		}
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		// sets err func (a *app) setup()
		if err = a.minifierConfig.BaseURL().Set(baseURL); err != nil {
			return
		}
		a.minifierConfig.BaseURL().Source = "ENV"
	}
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		if err = a.httpConfig.Address().Set(serverAddress); err != nil {
			return
		}
		a.httpConfig.Address().Source = "ENV"
	}

	// sets flags
	flag.Var(a.fileConfig.Path(), "f", fmt.Sprintf("Default: `%s`. Set the file path for the file storage", file.DefaultPath))
	flag.Var(a.minifierConfig.BaseURL(), "b", "Default: `http://localhost:8080`. Set the base URL for minified URLs")
	flag.Var(a.httpConfig.Address(), "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
	flag.Parse()

	if a.fileConfig.Path().String() == "" {
		a.Storager = memory.NewStrRecords()
	}
	a.Storager, err = file.New(&a.fileConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("storage set")

	a.Service = minifier.New(a.Storager, &a.minifierConfig)
	a.Server = http.NewServer(a.Service, &a.httpConfig)

	log.Printf("Base URL is set to `%s`", a.Service.Config.BaseURL().String())
	log.Printf("Server Address is set to `%s`", a.Server.Config.Address().String())

	return nil
}
