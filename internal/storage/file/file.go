package file

import (
	"encoding/json"
	"os"
	"sync"
)

func New() (*File, error) {
	err := os.Mkdir(path, dirPerm)
	if err != nil {
		return nil, err
	}

	fp, err := os.OpenFile(path+fileName, os.O_RDWR|os.O_CREATE, filePerm)
	if err != nil {
		return nil, err
	}

	return &File{
		p:       fp,
		encoder: json.NewEncoder(fp),
		decoder: json.NewDecoder(fp),
	}, nil
}

type File struct {
	mux     sync.RWMutex
	p       *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func (f *File) Close() error {
	return f.p.Close()
}

type fileRecord struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (f *File) Save(key, value string) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.p.Seek(0, 2) // set offset to the end of the file

	err := f.encoder.Encode(fileRecord{Key: key, Value: value})
	if err != nil {
		return err
	}

	return nil
}

func (f *File) Retrieve(key string) (value string, err error) {
	f.mux.RLock()
	defer f.mux.RUnlock()
	f.p.Seek(0, 0) // set offset to the start of the file

	var fr fileRecord
	for {
		err = f.decoder.Decode(&fr)
		if err != nil {
			return "", err
		}

		if fr.Key == key {
			return fr.Value, nil
		}
	}
}

// defaults
const (
	path                 = "./.files/"
	fileName             = "minifiedURLs.json"
	dirPerm  os.FileMode = 0o755 // UNIX persmiffions: Write Read Execute, Read _ Execute, Read _ Execute
	filePerm os.FileMode = 0o644 // UNIX persmiffions: Write Read _, Read _ _, Read _ _
)
