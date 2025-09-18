package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type result struct {
	code          int
	contentType   string
	contentLength string
	body          []byte
}

func Test_minifyURLHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		res  http.ResponseWriter
		req  *http.Request
		want result
		body string
	}{
		{
			name: "correct minify https://practicum.yandex.ru/",
			req:  httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("https://practicum.yandex.ru/"))),
			want: result{
				code:          201,
				contentType:   "text/plain",
				contentLength: "30",
				body:          []byte("http://example.com/DdGYF429IL4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			minifyURLHandler(w, tt.req)
			res := w.Result()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			defer res.Body.Close()
			assert.NoError(t, err)
			got := result{
				code:          res.StatusCode,
				contentType:   res.Header.Get("Content-Type"),
				contentLength: res.Header.Get("Content-Length"),
				body:          body,
			}
			require.Equal(t, tt.want, got)
		})
	}
}

type minifyURLResponse struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func Test_unMinifyURLHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req  *http.Request
		want minifyURLResponse
	}{
		{
			name: "correct unMinify https://practicum.yandex.ru/",
			req:  httptest.NewRequest("GET", "/DdGYF429IL4", nil),
			want: minifyURLResponse{
				statusCode: http.StatusTemporaryRedirect,
				headers: map[string]string{
					"Content-Length": "30",
					"Location":       "https://practicum.yandex.ru/",
				},
				body: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// preset the request and the environment
			id := tt.req.RequestURI[1:] // slice off leading "/"
			tt.req.SetPathValue("id", id)
			minifiedURLs[id] = tt.want.headers["Location"]

			rec := httptest.NewRecorder()
			unMinifyURLHandler(rec, tt.req)
			res := rec.Result()
			defer res.Body.Close()

			require.Equal(t, tt.want.statusCode, res.StatusCode)
			l, err := res.Location()
			require.NoError(t, err)

			require.Equal(t, tt.want.headers["Location"], l.String())
		})
	}
}
