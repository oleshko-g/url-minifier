package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
)

// Server is the internal implementation of [http.Server]
type Server struct {
	server *http.Server
	Service
	*Config
	logger
}

// Service is the expected URL minifier service
//
//go:generate moq -pkg minifier -out ../../mock/service/service.go . Service
type Service interface {
	MinifyURL(url string) (minifiedURL string, err error)
	UnMinifyURL(id string) (url string, err error)
	Ping() error
}

type logger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// NewServer configures and returns an internal [http.Server]
func NewServer(s Service, cp *Config) *Server {
	srv := &Server{
		Service: s,
		server:  &http.Server{},
		Config:  cp,
	}

	srv.Config.canDecompress = map[coding]struct{}{codingGZIP: {}}
	srv.Config.canCompress = []coding{codingGZIP, codingIdentity}

	r := chi.NewRouter()
	r.Use()
	r.Get("/ping", srv.withLoggingMiddleware(
		srv.pingHandler()))
	r.Post("/",
		srv.withLoggingMiddleware(
			srv.withEncodingMiddleware(
				srv.minifyURLHandler())))
	r.Get("/{id}",
		srv.withLoggingMiddleware(
			srv.withEncodingMiddleware(
				srv.unMinifyURLHandler())))
	r.Post("/api/shorten",
		srv.withLoggingMiddleware(
			srv.withEncodingMiddleware(
				srv.minifyURLJSONHandler())))
	srv.server.Handler = r

	srv.logger = slog.New(slog.Default().Handler())

	return srv
}

// ListenAndServe starts underlying [http.Server]
func (s *Server) ListenAndServe() error {
	s.server.Addr = s.Address().String()
	slog.Info(fmt.Sprintf("Minifier is listening on address: %s\n", s.server.Addr))
	return s.server.ListenAndServe()
}

// canDecompress reports if the server can decompress the given compression format
func (s *Server) canDecompress(compression coding) bool {
	_, ok := s.Config.canDecompress[compression]
	return ok
}

func (s *Server) chooseCompression(parsedAcceptCodings map[coding]qualityValue) (coding, error) {
	var compression parsedCoding

	for _, c := range s.canCompress {
		// "1. If no [Accept-Encoding] header field is in the request, any content coding is considered acceptable by the user agent."
		// [Accent-Encoding]: https://httpwg.org/specs/rfc9110.html#field.accept-encoding
		if parsedAcceptCodings == nil {
			compression.coding = c
			compression.qualityValue = 1.0
			break
		}

		// check if the client explicitly specified a coding which the server [canCompress]
		if q, ok := parsedAcceptCodings[c]; ok {

			if compression.qualityValue == 0 {
				compression.coding = c
			}

			if compression.coding == c {
				compression.qualityValue = q // the specified [qualityValue] overrides the default one
				continue
			}

			if compression.qualityValue < q {
				compression.coding = c
				compression.qualityValue = q
				continue
			}

		}

		// check if the client implicitly specified a coding which the server [canCompress]
		if q, ok := parsedAcceptCodings[codingWildcard]; ok {

			if compression.coding == c {
				compression.qualityValue = q // the specified [qualityValue] overrides the default one
				continue
			}

			if compression.qualityValue < q {
				compression.coding = c
				compression.qualityValue = q
				continue
			}

			if compression.qualityValue < q {
				compression.coding = c
				compression.qualityValue = q
				continue
			}

		}

	}

	if compression.qualityValue == 0 {
		return "", errNoCompressionChosen
	}

	return compression.coding, nil
}

// the client has forbidden every coding which the server can compress the response with
var errNoCompressionChosen = errors.New("the client has forbidden every coding which the server can compress a response with")

func (s *Server) pingHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		err := s.Service.Ping()
		if err != nil {
			responseWithError(res, err, http.StatusInternalServerError)
			s.logger.Error(err.Error())
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func (s *Server) minifyURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := validateContentType("text/plain", req.Header)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		data, err := io.ReadAll(req.Body)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		url, err := url.Parse(string(data))
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		s.logger.Debug(fmt.Sprintf("Original URL: %s", url))

		minifiedURL, err := s.Service.MinifyURL(url.String())
		if err != nil {
			responseWithError(res, err, http.StatusInternalServerError)
			s.logger.Error(err.Error())
			return
		}
		s.logger.Debug(fmt.Sprintf("minifiedURL: %s", minifiedURL))
		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Content-Length", strconv.Itoa(len(minifiedURL)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(minifiedURL))
	}
}

func (s *Server) unMinifyURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		id := req.PathValue("id")
		if id == "" {
			err = errors.New("empty id")
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}
		url, err := s.Service.UnMinifyURL(id)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		res.Header().Add("Location", url)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

type minifyURLRequest struct {
	URL string `json:"url"`
}
type minifyURLResponse struct {
	Result string `json:"result"`
}

func (s *Server) minifyURLJSONHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if err := validateContentType("application/json", req.Header); err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		// decode JSON request
		var reqBody minifyURLRequest
		d := json.NewDecoder(req.Body)
		if err := d.Decode(&reqBody); err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}
		defer req.Body.Close()

		// handle request
		minifiedURL, err := s.Service.MinifyURL(reqBody.URL)
		if err != nil {
			responseWithError(res, err, http.StatusInternalServerError)
			s.logger.Error(err.Error())
			return
		}

		// encode response
		resBody := minifyURLResponse{
			Result: minifiedURL,
		}
		jsonData, err := json.Marshal(&resBody)
		if err != nil {
			responseWithError(res, err, http.StatusInternalServerError)
			s.logger.Error(err.Error())
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(jsonData))
	}
}

func responseWithError(res http.ResponseWriter, err error, statusCode int) {
	http.Error(res, err.Error(), statusCode)
}

// validateContentType checks if the `mediaType` exists in the `headers`
func validateContentType(mediaType string, headers http.Header) error {
	parseMediaType, _, errParseMediaType := mime.ParseMediaType(headers.Get("Content-Type"))
	if errParseMediaType != nil {
		return errParseMediaType
	}

	if parseMediaType != mediaType {
		return errors.New("error: Content-Type isn't " + mediaType)
	}

	return nil
}
