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
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
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
	MinifyURLs(urls []map[string]string) (minifiedURLs []map[string]string, err error)
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
	r.Use(srv.withLoggingMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Use(srv.withEncodingMiddleware)
		r.Use(srv.withAuthorization)

		r.Post("/", srv.minifyURLHandler())
		r.Post("/api/shorten", srv.minifyURLJSONHandler())
		r.Post("/api/shorten/batch", srv.minifyURLsHandler())
	})
	r.Route("/api/user/urls", func(r chi.Router) {
		r.Use(srv.withEncodingMiddleware)
		r.Use(srv.withAuthentification)

		r.Get("/", srv.userURLsHandler())
	})

	r.Route("/{id}", func(r chi.Router) {
		r.Use(srv.withEncodingMiddleware)
		r.Get("/", srv.unMinifyURLHandler())
	})
	r.Get("/ping", srv.pingHandler())

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
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				responseWithError(res, err, http.StatusInternalServerError)
				s.logger.Error(err.Error())
				return
			}
			statusCode = http.StatusConflict
		}

		s.logger.Debug(fmt.Sprintf("minifiedURL: %s", minifiedURL))

		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Content-Length", strconv.Itoa(len(minifiedURL)))
		res.WriteHeader(statusCode)
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
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				responseWithError(res, err, http.StatusInternalServerError)
				s.logger.Error(err.Error())
				return
			}
			statusCode = http.StatusConflict
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
		res.WriteHeader(statusCode)
		res.Write([]byte(jsonData))
	}
}

// func (s *Server)

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

func (s *Server) minifyURLsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req minifyURLsRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			responseWithError(w, err, http.StatusBadRequest)
			return
		}

		var originalURLs []map[string]string
		for _, v := range req {
			originalURLs = append(originalURLs, v.toMap())
		}

		minifiedURLs, err := s.MinifyURLs(originalURLs)
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				responseWithError(w, err, http.StatusInternalServerError)
				s.logger.Error(err.Error())
				return
			}
			statusCode = http.StatusConflict
		}

		var resBody minifyURLsResponse
		for _, v := range minifiedURLs {
			var mURL minifyURLsResponseData
			mURL.fromMap(v)
			resBody = append(resBody, mURL)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err = json.NewEncoder(w).Encode(resBody); err != nil {
			s.logger.Error(err.Error())
		}
	}
}

type (
	minifyURLsRequest  []minifyURLsRequestData
	minifyURLsResponse []minifyURLsResponseData
)

type minifyURLsRequestData struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

func (m minifyURLsRequestData) toMap() map[string]string {
	return map[string]string{m.CorrelationID: m.OriginalURL}
}

func (m *minifyURLsResponseData) fromMap(ma map[string]string) {
	for i, v := range ma {
		m.CorrelationID = i
		m.ShortURL = v
	}
}

type minifyURLsResponseData struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s *Server) userURLsHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, _ = res, req
		res.WriteHeader(http.StatusNotImplemented)
	}
}
