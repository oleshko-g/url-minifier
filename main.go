package main //revive:disable-line

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/oleshko-g/url-minifier/internal/config"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage"
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	"github.com/oleshko-g/url-minifier/internal/storage/db/sql"
	"github.com/oleshko-g/url-minifier/internal/storage/file"
	"github.com/oleshko-g/url-minifier/internal/storage/memory"
	"github.com/oleshko-g/url-minifier/internal/transport/grpc"
	"github.com/oleshko-g/url-minifier/internal/transport/http"
)

var a app

func main() {
	printBuildInfo()
	if err := a.setup(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	go func() {
		cancel(a.http.Server.ListenAndServe())
		defer slog.Info(fmt.Sprintf("cause: %s", ctx.Err()))
		<-ctx.Done()
		a.http.Server.Shutdown(ctx)
		err := a.http.Server.Close()
		if err != nil {
			slog.Error("failed a graceful shutdown", "error", err.Error())
			return
		}
		slog.Info("shutdown the HTTP server gracefully")
	}()

	go func() {
		cancel(a.grpc.Server.ListenAndServe())
		<-ctx.Done()
		err := a.grpc.Server.GracefulShutdown()
		if err != nil {
			slog.Error("failed a graceful shutdown", "error", err.Error())
			return
		}
		slog.Info("shutdown the gRPC server gracefully")
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-shutdownSignal
	slog.Info(fmt.Sprintf("received the [%s] os.Signal. Shutting down gracefully...", sig.String()))
	cancel(errors.New("os signal: " + sig.String()))
}

type app struct {
	http struct {
		*http.Server
		*http.Config
	}
	grpc struct {
		*grpc.Server
		*grpc.Config
	}
	minifier struct {
		*minifier.Config
		*minifier.Service
	}
	storage.Storage
	sqlConfig      db.Config
	fileConfig     *file.Config
	configFilePath config.Option[*config.Path]
	configFile     *config.File
}

func (a *app) setup() (err error) {
	// Set the default config values
	a.configFilePath = config.NewPath()
	a.grpc.Config = grpc.NewConfig()
	a.http.Config = http.NewConfig()
	a.sqlConfig = db.NewConfig()
	a.fileConfig = file.NewConfig()
	a.minifier.Config = minifier.NewConfig()

	err = a.fileConfig.FilePath.Set(a.fileConfig.FilePath.Default)
	if err != nil {
		return err
	}
	err = a.minifier.Config.BaseURL.Set(a.minifier.Config.BaseURL.Default)
	if err != nil {
		return err
	}
	err = a.http.Config.Address.Set(a.http.Config.Address.Default)
	if err != nil {
		return err
	}

	// If an env var is present then it overrides the default value or the flag value
	godotenv.Load(".env")
	if cfgFilePath := os.Getenv(a.configFilePath.EnVarName); cfgFilePath != "" {
		if err = a.configFilePath.Set(cfgFilePath); err != nil {
			return err
		}
		a.configFilePath.Source = config.SourceEnv
	}
	if dbConn := os.Getenv(a.sqlConfig.DSN.EnVarName); dbConn != "" {
		// sets err func (a *app) setup()
		if err = a.sqlConfig.DSN.Set(dbConn); err != nil {
			return err
		}
		a.sqlConfig.DSN.Source = config.SourceEnv
	}
	if filePath := os.Getenv(a.fileConfig.FilePath.EnVarName); filePath != "" {
		// sets err func (a *app) setup()
		if err = a.fileConfig.FilePath.Set(filePath); err != nil {
			return err
		}
		a.fileConfig.FilePath.Source = config.SourceEnv
	}
	if trustedSubnet := os.Getenv(a.http.Config.TrustedIPSubnet.EnVarName); trustedSubnet != "" {
		if err = a.http.Config.TrustedIPSubnet.Set(trustedSubnet); err != nil {
			return err
		}
		a.http.Config.TrustedIPSubnet.Source = config.SourceEnv
	}
	if secured := os.Getenv(a.http.Config.Secured.EnVarName); secured != "" {
		if err = a.http.Config.Secured.Set(secured); err != nil {
			return err
		}
		a.http.Config.Secured.Source = config.SourceEnv
	}
	if serverAddress := os.Getenv(a.http.Config.Address.EnVarName); serverAddress != "" {
		if err = a.http.Config.Address.Set(serverAddress); err != nil {
			return err
		}
		a.http.Config.Address.Source = config.SourceEnv
	}

	if auditFile := os.Getenv(a.http.Config.AuditFile.EnVarName); auditFile != "" {
		if err = a.http.Config.AuditFile.Set(auditFile); err != nil {
			return err
		}
		a.http.Config.AuditFile.Source = config.SourceEnv
	}

	if auditURL := os.Getenv(a.http.Config.AuditURL.EnVarName); auditURL != "" {
		if err = a.http.Config.AuditURL.Set(auditURL); err != nil {
			return err
		}
		a.http.Config.AuditURL.Source = config.SourceEnv
	}
	if baseURL := os.Getenv(a.minifier.Config.BaseURL.EnVarName); baseURL != "" {
		if err = a.minifier.Config.BaseURL.Set(baseURL); err != nil {
			return err
		}
		a.minifier.Config.BaseURL.Source = config.SourceEnv
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
		fmt.Printf("%#v\n", a.configFile)

		// override only the defaults
		if a.configFile.BaseURL != "" && a.minifier.Config.BaseURL.Source == config.SourceDefault {
			a.minifier.Config.BaseURL.Set(a.configFile.BaseURL)
			a.minifier.Config.BaseURL.Source = config.SourceFile
		}
		if a.configFile.DatabaseDSN != "" && a.sqlConfig.DSN.Source == config.SourceDefault {
			a.sqlConfig.DSN.Set(a.configFile.DatabaseDSN)
			a.sqlConfig.DSN.Source = config.SourceFile
		}
		if a.configFile.FileStoragePath != "" && a.fileConfig.FilePath.Source == config.SourceDefault {
			a.fileConfig.FilePath.Set(a.configFile.FileStoragePath)
			a.fileConfig.FilePath.Source = config.SourceFile
		}
		if a.configFile.EnableHTTPS && a.http.Config.Secured.Source == config.SourceDefault {
			a.http.Config.Secured.Set(strconv.FormatBool(a.configFile.EnableHTTPS))
			a.http.Config.Secured.Source = config.SourceFile
		}
		if a.configFile.ServerAddress != "" && a.http.Config.Address.Source == config.SourceDefault {
			a.http.Config.Address.Set(a.configFile.ServerAddress)
			a.http.Config.Address.Source = config.SourceFile
		}
		if a.configFile.TrustedIPSubnet != "" && a.http.Config.TrustedIPSubnet.Source == config.SourceDefault {
			a.http.Config.TrustedIPSubnet.Set(a.configFile.TrustedIPSubnet)
			a.http.Config.TrustedIPSubnet.Source = config.SourceFile
		}

		return nil
	})
	flag.Var(a.sqlConfig.DSN, a.sqlConfig.DSN.Name, a.sqlConfig.DSN.Description)
	flag.Var(a.fileConfig.FilePath, a.fileConfig.FilePath.Name, a.fileConfig.FilePath.Description)
	flag.Var(a.http.Config.Secured, a.http.Config.Secured.Name, a.http.Config.Secured.Description)
	flag.Var(a.http.Config.TrustedIPSubnet, a.http.Config.TrustedIPSubnet.Name, a.http.Config.TrustedIPSubnet.Description)
	flag.Var(a.http.Config.Address, a.http.Config.Address.Name, a.http.Config.Address.Description)
	flag.Var(a.http.Config.AuditFile, a.http.Config.AuditFile.Name, a.http.Config.AuditFile.Description)
	flag.Var(a.http.Config.AuditURL, a.http.Config.AuditURL.Name, a.http.Config.AuditURL.Description)
	flag.Var(a.minifier.Config.BaseURL, a.minifier.Config.BaseURL.Name, a.minifier.Config.BaseURL.Description)
	// If flags are present then [flag.Parse] overrides defaults or env vars.
	flag.Parse()

	if a.sqlConfig.DSN.String() != "" {
		sqlStorage, err := sql.New(a.sqlConfig)
		if err != nil {
			return err
		}
		a.Storage.Storager = sqlStorage
		a.Storage.Pinger = sqlStorage
		a.Storage.Closer = sqlStorage
		a.Storage.Counter = sqlStorage
		slog.Info("The storage is set to db.")
	} else if a.fileConfig.FilePath.String() != "" {
		fileStorage, err := file.New(a.fileConfig)
		if err != nil {
			return err
		}
		a.Storage.Storager = fileStorage
		a.Storage.Closer = fileStorage
		slog.Info("The storage is set to file.")
	} else {
		a.Storage.Storager = memory.NewStrRecords()
		slog.Info("The storage is set to memory.")
	}

	if err != nil {
		return err
	}

	a.minifier.Service = minifier.New(&a.Storage, a.minifier.Config)
	a.http.Server = http.NewServer(a.minifier.Service, a.http.Config)
	a.grpc.Server = grpc.NewServer(a.grpc.Config, a.minifier.Service)

	slog.Info(fmt.Sprintf("Base URL is set to `%s`", a.minifier.Service.Config.BaseURL.String()))
	slog.Info(fmt.Sprintf("Server Address is set to `%s`", a.http.Server.Config.Address.String()))

	return nil
}
