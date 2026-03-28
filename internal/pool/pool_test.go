package pool_test

import (
	"testing"

	"github.com/oleshko-g/url-minifier/cmd/reset/example"
	"github.com/oleshko-g/url-minifier/internal/pool"
)

func Test_pool(t *testing.T) {
	p := pool.New[*example.GenStruct]()

	t.Run("Get", func(t *testing.T) {
		p.Get()
	})
	t.Run("Put", func(t *testing.T) {
		p.Put(&example.GenStruct{})
	})
}
