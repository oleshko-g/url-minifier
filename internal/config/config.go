// Package config provides a generic config [Option]
package config

import (
	"flag"
)

type Option[T flag.Value] struct {
	Name        string
	Value       T
	Description string
	Default     string
	Source
}

type Source int

const (
	SourceDefault Source = iota
	SourceFile
	SourceEnv
	SourceFlag
)

func (receiver Option[T]) Set(s string) error {
	return receiver.Value.Set(s)
}

func (receiver Option[T]) String() string {
	return receiver.Value.String()
}

func NewPath() Option[*Path] {
	return Option[*Path]{
		Name:        "c",
		Value:       new(Path),
		Description: "The path to the config file which is applied first and then gets overridden by flags or env vars",
		Default:     "",
		Source:      SourceDefault,
	}
}

type Path string

func (p Path) String() string {
	return string(p)
}

func (p Path) Set(s string) error {
	p = Path(s)
	return nil
}

type File struct {
	ServerAddress   string `json:"server_address" default:"localhost:8080"` // SERVER_ADDRESS or flag -a
	BaseURL         string `json:"base_url" default:"http://localhost"`     // BASE_URL or flag -b
	FileStoragePath string `json:"file_storage_path" default:""`            // FILE_STORAGE_PATH flag -f
	DatabaseDSN     string `json:"database_dsn" default:""`                 // DATABASE_DSN or flag -d
	EnableHTTPS     string `json:"enable_https" default:""`                 // ENABLE_HTTPS or flag -s
}
