package file

import (
	"os"

	"github.com/rs/zerolog"
)

func New() *File {
	zl := zerolog.New(os.Stderr).With().Timestamp().Logger()
	file := File{
		logger: zl,
	}

	fp, err := os.OpenFile(defaultPath+defaultFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, defaultPermissions)
	if err != nil {
		file.logger.Fatal().Msg(err.Error())
	}
	file.p = fp

	return &file
}

type File struct {
	p      *os.File
	logger zerolog.Logger
}

func (f *File) Save(key, value string) error {
	return nil
}

func (f *File) Retrieve(key string) (value string, err error) {
	return "", nil
}

const (
	defaultPath                    = "./.files/"
	defaultFileName                = "minifiedURLs.json"
	defaultPermissions os.FileMode = 0o644 // UNIX persmiffions: WriteRread_, _Read_, _Read_
)
