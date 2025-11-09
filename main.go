package main //revive:disable-line

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	"github.com/oleshko-g/url-minifier/internal/storage/file"
	"github.com/oleshko-g/url-minifier/internal/storage/memory"
	"github.com/oleshko-g/url-minifier/internal/transport/http"
)

var a app

func main() {
	if err := a.setup(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Error(a.Server.ListenAndServe().Error())
}

type app struct {
	sqlConfig      db.Config
	fileConfig     file.Config
	minifierConfig minifier.Config
	httpConfig     http.Config
	minifier.Storager
	*minifier.Service
	*http.Server
}

func (a *app) setup() (err error) {
	// Set the default config values

	err = a.fileConfig.Path().Set(string(file.DefaultFilepath))
	if err != nil {
		return err
	}
	err = a.minifierConfig.BaseURL().Set("http://localhost:8080")
	if err != nil {
		return err
	}
	err = a.httpConfig.Address().Set("localhost:8080")
	if err != nil {
		return err
	}
	a.minifierConfig.MaxLen = 8

	// If an env var is present then it overrides the default value or the flag value
	godotenv.Load(".env")
	if dbConn := os.Getenv("DATABASE_DSN"); dbConn != "" {
		// sets err func (a *app) setup()
		if err = a.sqlConfig.DSN().Set(dbConn); err != nil {
			return err
		}
	}
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		// sets err func (a *app) setup()
		if err = a.fileConfig.Path().Set(filePath); err != nil {
			return err
		}
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		// sets err func (a *app) setup()
		if err = a.minifierConfig.BaseURL().Set(baseURL); err != nil {
			return err
		}
		a.minifierConfig.BaseURL().Source = "ENV"
	}
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		if err = a.httpConfig.Address().Set(serverAddress); err != nil {
			return err
		}
		a.httpConfig.Address().Source = "ENV"
	}

	// sets flags
	flag.Var(a.sqlConfig.DSN(), "d", "Set the sql db connection string")
	flag.Var(a.fileConfig.Path(), "f", fmt.Sprintf("Default: `%s`. Set the file path for the file storage", file.DefaultFilepath))
	flag.Var(a.minifierConfig.BaseURL(), "b", "Default: `http://localhost:8080`. Set the base URL for minified URLs")
	flag.Var(a.httpConfig.Address(), "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
	flag.Parse()

	if a.fileConfig.Path().String() == "" {
		a.Storager = memory.NewStrRecords()
		slog.Info("The storage is set to memory.")
	} else {
		a.Storager, err = file.New(&a.fileConfig)
		slog.Info("The storage is set to file.")
	}
	if err != nil {
		return err
	}

	a.Service = minifier.New(a.Storager, &a.minifierConfig)
	a.Server = http.NewServer(a.Service, &a.httpConfig)

	slog.Info(fmt.Sprintf("Base URL is set to `%s`", a.Service.Config.BaseURL().String()))
	slog.Info(fmt.Sprintf("Server Address is set to `%s`", a.Server.Config.Address().String()))

	return nil
}
