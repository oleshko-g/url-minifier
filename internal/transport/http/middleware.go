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
			s.logger.Error(err.Error())
			return
		}

		// iterate through Content-Encoding's parsedCodings and---if the server can---decompress each one
		for _, v := range parsedConentCodings {
			if !s.canDecompress(v.coding) {
				err := fmt.Errorf("the server can't decompress the %v coding", v.coding)
				http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
				s.logger.Error(err.Error())
				return
			}

			switch v.coding {
			case codingGZIP:
				gzr, err := gzip.NewReader(req.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					s.logger.Error(err.Error())
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
			s.logger.Error(err.Error())
			return
		}

		compression, err := s.chooseCompression(parsedAcceptCodings)
		if err != nil {
			if errors.Is(err, errNoCompressionChosen) {
				w.Header().Set("Accept-Encoding", string(s.canCompress.String()))
			}
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			s.logger.Error(err.Error())
			return
		}
		cw := compressingResponseWriter{
			ResponseWriter: w,
			coding:         compression,
			compressor:     nil, // can't set compressor right away because it depends on the "Content-Type" of an HTTP response
		}

		// if the handler choose the compressor then compressor must be closed
		defer func() {
			if cw.compressor != nil {
				cw.compressor.Close()
			}
		}()

		h(&cw, req)
	}
}

type compressingResponseWriter struct {
	http.ResponseWriter
	compressor writeFlushCloser
	coding
}

type writeFlushCloser interface {
	io.WriteCloser
	Flush() error
}

// WriteHeader chooses a compressor based on the "Content-Type" of [http.ResponseWriter]
func (cw *compressingResponseWriter) WriteHeader(statusCode int) {
	cw.chooseCompressor()
	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *compressingResponseWriter) Write(reponseBody []byte) (n int, err error) {
	if cw.compressor != nil {
		return cw.write(reponseBody)
	}
	return cw.ResponseWriter.Write(reponseBody)
}

// chooseCompressor might set the compressor and "Content-Encoding" of [http.ResponseWriter] based on its "Content-Type" [http.Header]
func (cw *compressingResponseWriter) chooseCompressor() {
	switch cw.ResponseWriter.Header().Get("Content-Type") {
	case "application/json", "text/html":
		switch cw.coding {
		case codingGZIP:
			cw.ResponseWriter.Header().Set("Content-Encoding", string(cw.coding))
			cw.ResponseWriter.Header().Del("Content-Length") // compresstion will change the content-length which a handler might've set
			cw.compressor = gzip.NewWriter(cw.ResponseWriter)
		case codingIdentity:
			// no compression
		}
	}
}

func (cw *compressingResponseWriter) write(dataToCompress []byte) (n int, err error) {
	if cw.compressor != nil {
		n, err = cw.compressor.Write(dataToCompress)
		if err != nil {
			err = cw.compressor.Flush() // need to close to Flush compressor's in memory buffer
			return n, err
		}
		return n, err
	}
	return 0, errors.New("compressor is nil")
}

func (s *Server) withLoggingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}

		h(&lw, req)

		s.logger.Info("Request:",
			"URI", req.RequestURI,
			"Method", req.Method,
			"Duration, ns", time.Since(start),
		)

		s.logger.Info("Response:",
			"Status Code", lw.statusCode,
			"Content size, bytes", lw.contentLength,
		)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (lw *loggingResponseWriter) Write(responseBody []byte) (n int, err error) {
	n, err = lw.ResponseWriter.Write(responseBody)
	lw.contentLength += n
	return n, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.statusCode = statusCode
}
