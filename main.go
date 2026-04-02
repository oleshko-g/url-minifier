package main //revive:disable-line

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/oleshko-g/url-minifier/internal/config"
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
	configFilePath config.Option[*config.Path]
	configFile     *config.File
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
	a.configFilePath = config.NewPath()
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
	if cfgFilePath := os.Getenv("CONFIG"); cfgFilePath != "" {
		if err = a.configFilePath.Set(cfgFilePath); err != nil {
			return err
		}
		a.configFilePath.Source = config.SourceEnv
	}
	if dbConn := os.Getenv("DATABASE_DSN"); dbConn != "" {
		// sets err func (a *app) setup()
		if err = a.sqlConfig.DSN.Set(dbConn); err != nil {
			return err
		}
		a.sqlConfig.DSN.Source = config.SourceEnv
	}
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		// sets err func (a *app) setup()
		if err = a.fileConfig.FilePath.Set(filePath); err != nil {
			return err
		}
		a.fileConfig.FilePath.Source = config.SourceEnv
	}
	if secured := os.Getenv("ENABLE_HTTPS"); secured != "" {
		if err = a.httpConfig.Secured.Set(secured); err != nil {
			return err
		}
		a.httpConfig.Secured.Source = config.SourceEnv
	}
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		if err = a.httpConfig.Address.Set(serverAddress); err != nil {
			return err
		}
		a.httpConfig.Address.Source = config.SourceEnv
	}

	if auditFile := os.Getenv("AUDIT_FILE"); auditFile != "" {
		if err = a.httpConfig.AuditFile.Set(auditFile); err != nil {
			return err
		}
		a.httpConfig.AuditFile.Source = config.SourceEnv
	}

	if auditURL := os.Getenv("AUDIT_URL"); auditURL != "" {
		if err = a.httpConfig.AuditFile.Set(auditURL); err != nil {
			return err
		}
		a.httpConfig.AuditURL.Source = config.SourceEnv
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		if err = a.minifierConfig.BaseURL.Set(baseURL); err != nil {
			return err
		}
		a.minifierConfig.BaseURL.Source = config.SourceEnv
	}

	// set flags
	flag.Func(a.configFilePath.Name, a.configFilePath.Description, func(s string) error {
		if err := a.configFilePath.Set(s); err != nil {
			return err
		}

		cfgFile, err := os.Open(s)
		if err != nil {
			return err
		}

		d := json.NewDecoder(cfgFile)
		if err := d.Decode(&a.configFile); err != nil {
			return err
		}
		fmt.Printf("%v\n", a.configFile)

		// override only the defaults
		if a.configFile.BaseURL != "" && a.minifierConfig.BaseURL.Source == config.SourceDefault {
			a.minifierConfig.BaseURL.Set(a.configFile.BaseURL)
		}
		if a.configFile.DatabaseDSN != "" && a.sqlConfig.DSN.Source == config.SourceDefault {
			a.sqlConfig.DSN.Set(a.configFile.DatabaseDSN)
		}
		if a.configFile.FileStoragePath != "" && a.fileConfig.FilePath.Source == config.SourceDefault {
			a.fileConfig.FilePath.Set(a.configFile.FileStoragePath)
		}
		if a.configFile.EnableHTTPS && a.httpConfig.Secured.Source == config.SourceDefault {
			a.httpConfig.Secured.Set(strconv.FormatBool(a.configFile.EnableHTTPS))
		}
		if a.configFile.ServerAddress != "" && a.httpConfig.Address.Source == config.SourceDefault {
			a.httpConfig.Address.Set(a.configFile.ServerAddress)
		}

		return nil
	})
	flag.Var(a.sqlConfig.DSN, a.sqlConfig.DSN.Name, a.sqlConfig.DSN.Description)
	flag.Var(a.fileConfig.FilePath, a.fileConfig.FilePath.Name, a.fileConfig.FilePath.Description)
	flag.Var(a.httpConfig.Secured, a.httpConfig.Secured.Name, a.httpConfig.Secured.Description)
	flag.Var(a.httpConfig.Address, a.httpConfig.Address.Name, a.httpConfig.Address.Description)
	flag.Var(a.httpConfig.AuditFile, a.httpConfig.AuditFile.Name, a.httpConfig.AuditFile.Description)
	flag.Var(a.httpConfig.AuditURL, a.httpConfig.AuditURL.Name, a.httpConfig.AuditURL.Description)
	flag.Var(a.minifierConfig.BaseURL, a.minifierConfig.BaseURL.Name, a.minifierConfig.BaseURL.Description)
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
