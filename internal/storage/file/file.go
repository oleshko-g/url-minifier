package file

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// New returns a pointer to a [File] or an error. If [path] is empty it sets it to [DefaultPath]
func New(c *Config) (*File, error) {
	file := File{
		mux:    sync.RWMutex{},
		Config: c,
	}

	err := os.MkdirAll(c.Path().String(), dirPerm)
	if err != nil {
		return nil, err
	}

	fp, err := os.OpenFile(c.Path().String()+string(os.PathSeparator)+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}
	file.p = fp

	return &file, nil
}

type File struct {
	mux sync.RWMutex
	p   *os.File
	*Config
}

func (f *File) Close() error {
	return f.p.Close()
}

type record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (f *File) Save(key, value string) (err error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	defer f.p.Sync()

	return json.NewEncoder(f.p).Encode(record{Key: key, Value: value})
}

func (f *File) Retrieve(key string) (value string, err error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	rr, err := f.newRecordReader()
	if err != nil {
		return "", err
	}
	defer rr.file.Close()

	var fr record
	for {
		err = rr.decoder.Decode(&fr)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", storageErrors.ErrNotFound
			}
			return "", err
		}

		if fr.Key == key {
			return fr.Value, nil
		}
	}
}

type recordReader struct {
	file    *os.File
	decoder *json.Decoder
}

func (f *File) newRecordReader() (*recordReader, error) {
	fp, err := os.OpenFile(f.Config.Path().String()+fileName, os.O_RDONLY|os.O_CREATE, filePerm)
	if err != nil {
		return nil, err
	}

	dep := json.NewDecoder(fp)

	return &recordReader{
		file:    fp,
		decoder: dep,
	}, nil
}

// defaults
const (
	DefaultPath = "./files/"
	fileName    = "minifiedURLs.json"
)

// UNIX persmissions
const (
	// Write Read Execute, Read _ Execute, Read _ Execute
	dirPerm os.FileMode = 0o755
	// Write Read _, Read _ _, Read _ _
	filePerm os.FileMode = 0o644
)
