package main //revive:disable-line

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage"
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	"github.com/oleshko-g/url-minifier/internal/storage/db/sql"
	"github.com/oleshko-g/url-minifier/internal/storage/file"
	"github.com/oleshko-g/url-minifier/internal/storage/memory"
	"github.com/oleshko-g/url-minifier/internal/transport/http"
)

var a app

func main() {
	printBuildInfo()
	if err := a.setup(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Error(a.Server.ListenAndServe().Error())
}

type app struct {
	sqlConfig      db.Config
	fileConfig     *file.Config
	minifierConfig *minifier.Config
	httpConfig     *http.Config
	storage.StoragePinger
	*minifier.Service
	*http.Server
}

func (a *app) setup() (err error) {
	// Set the default config values
	a.httpConfig = http.NewConfig()
	a.sqlConfig = db.NewConfig()
	a.fileConfig = file.NewConfig()
	a.minifierConfig = minifier.NewConfig()

	err = a.fileConfig.FilePath.Set(a.fileConfig.FilePath.Default)
	if err != nil {
		return err
	}
	err = a.minifierConfig.BaseURL.Set(a.minifierConfig.BaseURL.Default)
	if err != nil {
		return err
	}
	err = a.httpConfig.Address.Set(a.httpConfig.Address.Default)
	if err != nil {
		return err
	}

	// If an env var is present then it overrides the default value or the flag value
	godotenv.Load(".env")
	if dbConn := os.Getenv("DATABASE_DSN"); dbConn != "" {
		// sets err func (a *app) setup()
		if err = a.sqlConfig.DSN.Set(dbConn); err != nil {
			return err
		}
	}
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		// sets err func (a *app) setup()
		if err = a.fileConfig.FilePath.Set(filePath); err != nil {
			return err
		}
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		// sets err func (a *app) setup()
		if err = a.minifierConfig.BaseURL.Set(baseURL); err != nil {
			return err
		}
		a.minifierConfig.BaseURL.Source = "ENV"
	}
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		if err = a.httpConfig.Address.Set(serverAddress); err != nil {
			return err
		}
		a.httpConfig.Address.Source = "ENV"
	}

	if auditFile := os.Getenv("AUDIT_FILE"); auditFile != "" {
		if err = a.httpConfig.AuditFile.Set(auditFile); err != nil {
			return err
		}
		a.httpConfig.AuditFile.Source = "ENV"
	}

	if auditURL := os.Getenv("AUDIT_URL"); auditURL != "" {
		if err = a.httpConfig.AuditFile.Set(auditURL); err != nil {
			return err
		}
		a.httpConfig.AuditURL.Source = "ENV"
	}

	// set flags
	flag.Var(a.sqlConfig.DSN, a.sqlConfig.DSN.Name, a.sqlConfig.DSN.Description)
	flag.Var(a.fileConfig.FilePath, a.fileConfig.FilePath.Name, a.fileConfig.FilePath.Description)
	flag.Var(a.minifierConfig.BaseURL, a.minifierConfig.BaseURL.Name, a.minifierConfig.BaseURL.Description)
	flag.Var(a.httpConfig.Address, a.httpConfig.Address.Name, a.httpConfig.Address.Description)
	flag.Var(a.httpConfig.AuditFile, a.httpConfig.AuditFile.Name, a.httpConfig.AuditFile.Description)
	flag.Var(a.httpConfig.AuditURL, a.httpConfig.AuditURL.Name, a.httpConfig.AuditURL.Description)
	flag.Var(a.httpConfig.Secured, a.httpConfig.Secured.Name, a.httpConfig.Secured.Description)
	// If flags are present then [flag.Parse] overrides defaults or env vars.
	flag.Parse()

	if a.sqlConfig.DSN.String() != "" {
		a.StoragePinger, err = sql.New(a.sqlConfig)
		slog.Info("The storage is set to db.")
	} else if a.fileConfig.FilePath.String() != "" {
		a.StoragePinger, err = file.New(a.fileConfig)
		slog.Info("The storage is set to file.")
	} else {
		a.StoragePinger = memory.NewStrRecords()
		slog.Info("The storage is set to memory.")
	}

	if err != nil {
		return err
	}

	a.Service = minifier.New(a.StoragePinger, a.minifierConfig)
	a.Server = http.NewServer(a.Service, a.httpConfig)

	slog.Info(fmt.Sprintf("Base URL is set to `%s`", a.Service.Config.BaseURL.String()))
	slog.Info(fmt.Sprintf("Server Address is set to `%s`", a.Server.Config.Address.String()))

	return nil
}
