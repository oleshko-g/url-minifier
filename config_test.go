package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newMinifierConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		args []string
		want config
	}{
		{
			name: "address: localhost:9000",
			args: []string([]string{"-a", "localhost:9000"}),
			want: config{
				a:      address{host: "localhost", port: "9000"},
				b:      defaultConfig.b,
				maxLen: defaultConfig.maxLen,
			},
		},
		{
			name: "baseURL: https://localhost:9000",
			args: []string([]string{"-b", "https://localhost:9000"}),
			want: config{
				a:      defaultConfig.a,
				b:      baseURL{scheme: "https", address: address{host: "localhost", port: "9000"}},
				maxLen: defaultConfig.maxLen,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig
			fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			fs.Var(&cfg.a, "a", "Default: `localhost:8080`. Sets the network address and the port for the minifier")
			fs.Var(&cfg.b, "b", "Default: `https://localhost:8080`. Set the base URL for minified URLs")
			fs.Parse(tt.args)
			got := cfg
			assert.Equal(t, tt.want, got)
		})
	}
}
