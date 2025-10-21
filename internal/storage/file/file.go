package file

import "os"

type File struct {
	*os.File
}

func (f *File) Save(key, value string) error {
	return nil
}

func (f *File) Retrieve(key string) (value string, err error) {
	return "", nil
}
