package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMinifierResponse struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func TestServer_minifyURLHandler(t *testing.T) {
	storage := memory.NewStrRecords()
	service := minifier.New(storage)
	service.Config.MaxLen = 8
	service.Config.BaseURL().Set("http://localhost:8080/")
	server := NewServer(service)

	tests := []struct {
		name        string // description of this test case
		originalURL string
		minifiedID  string
		want        testMinifierResponse
	}{
		{
			name:        "correct minify https://practicum.yandex.ru/",
			originalURL: "https://practicum.yandex.ru/",
			minifiedID:  "DdGYF429IL4",
			want: testMinifierResponse{
				statusCode: 201,
				headers: map[string]string{
					"Content-Type":   "text/plain",
					"Content-Length": strconv.Itoa(len(service.BaseURL().String()) + 12),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.body = []byte(service.Config.BaseURL().String() + "/" + tt.minifiedID)

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("https://practicum.yandex.ru/")))
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			server.minifyURLHandler().ServeHTTP(w, req)
			res := w.Result()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, tt.want.statusCode, res.StatusCode)

			for k, v := range tt.want.headers {
				require.Equal(t, v, res.Header.Get(k), "Asserting %s: %s", k, v)
			}

			require.Equal(t, tt.want.body, body)
		})
	}
}

func TestServer_unMinifyURLHandler(t *testing.T) {
	storage := memory.NewStrRecords()
	service := minifier.New(storage)
	service.Config.MaxLen = 8
	service.Config.BaseURL().Set("http://localhost:8080/")
	server := NewServer(service)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		originalURL string
		minifiedID  string
		want        testMinifierResponse
	}{
		{
			name:        "correct unMinify https://practicum.yandex.ru/",
			minifiedID:  "DdGYF429IL4",
			originalURL: "https://practicum.yandex.ru/",
			want: testMinifierResponse{
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
			storage.Save(tt.minifiedID, tt.originalURL)

			req := httptest.NewRequest("GET", "/"+tt.minifiedID, nil)
			req.SetPathValue("id", tt.minifiedID)

			w := httptest.NewRecorder()
			server.unMinifyURLHandler().ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.want.statusCode, res.StatusCode)
			l, err := res.Location()
			require.NoError(t, err)

			require.Equal(t, tt.want.headers["Location"], l.String())
		})
	}
}

func TestServer_minifyJSONURLHandler(t *testing.T) {
	// setup server config to use it programmatically
	storage := memory.NewStrRecords()
	service := minifier.New(storage)
	service.Config.MaxLen = 8
	service.Config.BaseURL().Set("http://localhost:8080/")
	server := NewServer(service)

	tests := []struct {
		name        string // description of this test case
		originalURL string
		minifiedID  string
		want        testMinifierResponse
	}{
		{
			name:        "Correct minify 'https://practicum.yandex.ru/'",
			originalURL: "https://practicum.yandex.ru/",
			minifiedID:  "DdGYF429IL4",
			want: testMinifierResponse{
				statusCode: 201,
				headers: map[string]string{
					"Content-Type":   "application/json",
					"Content-Length": "46",
				},
				body: make([]byte, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set wanted response body
			resData, err := json.Marshal(minifyURLResponse{Result: service.BaseURL().String() + "/" + tt.minifiedID})
			require.NoError(t, err)
			tt.want.body = resData

			// set up request
			reqData, err := json.Marshal(minifyURLRequest{URL: tt.originalURL})
			require.NoError(t, err)
			b := bytes.NewBuffer(reqData)
			req := httptest.NewRequest("POST", "/shorten/", b)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(b.Len()))

			// make the request
			w := httptest.NewRecorder()
			server.minifyURLJSONHandler().ServeHTTP(w, req)
			res := w.Result()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			defer res.Body.Close()

			// tests
			require.Equal(t, tt.want.statusCode, res.StatusCode)

			for k, v := range tt.want.headers {
				require.Equal(t, v, res.Header.Get(k), "Asserting %s: %s", k, v)
			}

			require.Equal(t, tt.want.body, body)
		})
	}
}

func TestServer_chooseCompression(t *testing.T) {
	type chooseCompressionResult struct {
		coding
		error
	}
	tests := []struct {
		name                string // description of this test case
		s                   service
		parsedAcceptCodings map[coding]qualityValue
		want                chooseCompressionResult
	}{
		{
			name: "the client specified GZIP",
			s:    nil,
			parsedAcceptCodings: map[coding]qualityValue{
				codingGZIP: 1.0,
			},
			want: chooseCompressionResult{coding: codingGZIP, error: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(tt.s)
			got, gotErr := s.chooseCompression(tt.parsedAcceptCodings)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want.coding, got)
		})
	}
}
