package http

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/stretchr/testify/require"
)

type testMinifierResponse struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func TestServer_minifyURLHandler(t *testing.T) {
	mockService := minifier.NewMockMinifier()
	mockService.Config.MaxLen = 8
	mockService.Config.BaseURL().Set("http://localhost:8080/")
	server := NewServer(mockService)

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
					"Content-Length": strconv.Itoa(len(mockService.BaseURL().String()) + 12),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.OriginalURLs[tt.originalURL] = tt.minifiedID
			mockService.MinifiedIDs[tt.minifiedID] = tt.originalURL
			tt.want.body = []byte(mockService.Config.BaseURL().String() + "/" + tt.minifiedID)

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
