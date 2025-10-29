// Package file is an implementation of [minifier.Storager]
package file

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		configp  *Config
		wantFile *File
		wantErr  error
	}{
		{
			name:     "no such file or directory",
			configp:  &Config{fpath: ""},
			wantFile: nil,
			wantErr:  fs.ErrNotExist,
		},
		{
			name:     "default file path",
			configp:  &Config{fpath: DefaultFilepath},
			wantFile: &File{},
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, err := New(tt.configp)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantFile != nil {
				assert.NotNil(t, gotFile, tt.wantFile)
			}
		})
	}
}
