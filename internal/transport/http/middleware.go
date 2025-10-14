package http

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s *Server) withEncodingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		parsedCodings, err := parseContentEncoding(req.Header.Values("Content-Encoding"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// iterate through Content-Encoding's parsedCodings and---if the server can---decompress each
		for _, v := range parsedCodings {
			if !s.canDecompress(v.coding) {
				http.Error(w, fmt.Errorf("the server can't decompress the %v coding", v.coding).Error(), http.StatusUnsupportedMediaType)
				return
			}

			switch v.coding {
			case codingGZIP:
				gzr, err := gzip.NewReader(req.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer gzr.Close()

				req.Body = gzr // replace the request body with decompressor for reading the next decompressor or regular HTTP request handler
			}
		}

		var chosenCompression parsedCoding
		parsedAcceptCodings, err := parseAcceptEncoding(req.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for k := range s.Config.canCompress {
			if q, ok := parsedAcceptCodings[k]; ok {
				chosenCompression = parsedCoding.coding
			}
		}

		h(w, req)
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (s *Server) withLoggingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}

		h(&lw, req)

		// log requst
		s.Logger.Info().
			Str("URI", req.RequestURI).
			Str("Method", req.Method).
			Dur("Duration, ns", time.Since(start)).
			Msg("")

		// log response
		s.Logger.Info().
			Int("Status Code", lw.statusCode).
			Int("Content size, bytes", lw.contentLength).
			Msg("")
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (lw *loggingResponseWriter) Write(data []byte) (n int, err error) {
	n, err = lw.ResponseWriter.Write(data)
	lw.contentLength = n
	if err != nil {
		return 0, err
	}

	return
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.statusCode = statusCode
}
