package http

import (
	"net/http"
	"time"
)

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
