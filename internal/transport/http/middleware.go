package http

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s *Server) withEncodingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		parsedConentCodings, err := parseContentEncoding(req.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.Logger.Err(err).Msg("")
			return
		}

		// iterate through Content-Encoding's parsedCodings and---if the server can---decompress each
		for _, v := range parsedConentCodings {
			if !s.canDecompress(v.coding) {
				err := fmt.Errorf("the server can't decompress the %v coding", v.coding)
				http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
				s.Logger.Err(err).Msg("")
				return
			}

			switch v.coding {
			case codingGZIP:
				gzr, err := gzip.NewReader(req.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					s.Logger.Err(err).Msg("")
					return
				}
				defer gzr.Close()

				req.Body = gzr // replace the request body with decompressor for reading the next decompressor or regular HTTP request handler
			}
		}

		// parse the Accept-Encdoing(s) the client could specify in the request and---if the server can---compress the response
		parsedAcceptCodings, err := parseAcceptEncoding(req.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.Logger.Err(err).Msg("")
			return
		}

		compression, err := s.chooseCompression(parsedAcceptCodings)
		if err != nil {
			if errors.Is(err, errNoCompressionChosen) {
				w.Header().Set("Accept-Encoding", string(s.canCompress.String()))
			}
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			s.Logger.Err(err).Msg("")
			return
		}
		cw := compressingResponseWriter{
			ResponseWriter: w,
			coding:         compression,
		}

		h(cw, req)
	}
}

type compressingResponseWriter struct {
	http.ResponseWriter
	io.WriteCloser
	coding
}

func (cw compressingResponseWriter) Write(b []byte) (int, error) {
	switch cw.ResponseWriter.Header().Get("Content-Type") {
	case "application/json", "text/html":

		switch cw.coding {
		case codingGZIP:
			cw.WriteCloser = gzip.NewWriter(cw.ResponseWriter)
		case codingIdentity:
			return cw.ResponseWriter.Write(b) // write without compression
		}
		if cw.WriteCloser != nil {
			defer cw.WriteCloser.Close()
		}

		cw.ResponseWriter.Header().Set("Content-Encoding", string(cw.coding))
		return cw.ResponseWriter.Write(b)
	}

	return cw.ResponseWriter.Write(b)
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
