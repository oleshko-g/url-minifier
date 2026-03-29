package http //revive:disable-line:var-naming

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/http/pprof"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
)

// Server is the internal implementation of [http.Server]
type Server struct {
	server *http.Server
	Service
	*Config
	logger
	auditors      []auditor
	auditSubjects []chan auditEvent
}

// Service is the expected URL minifier service
//
//go:generate moq -pkg minifier -out ../../mock/service/service.go . Service
type Service interface {
	MinifyURL(ctx context.Context, userID string, url string) (minifiedURL string, err error)

	MinifyURLs(ctx context.Context, userID string,
		urls []map[string]string) (minifiedURLs []map[string]string, err error)
	UnMinifyURL(id string) (url string, err error)
	UserURLs(ctx context.Context, userID string) ([]minifier.URL, error)
	// DeleteUserURLs deletes a batch of shortened URLs
	DeleteUserURLs(ctx context.Context, userID string, minifiedIDs []string) error
	UnMinifyUserURL(ctx context.Context, id string) (url string, isDeleted bool, err error)
	Ping() error
}

type logger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// NewServer configures and returns an internal [http.Server]
func NewServer(s Service, cfg *Config) *Server {
	srv := &Server{
		Service: s,
		server:  &http.Server{},
		Config:  cfg,
	}
	srv.Config.canDecompress = map[coding]struct{}{codingGZIP: {}}
	srv.Config.canCompress = []coding{codingGZIP, codingIdentity}

	if *srv.Config.Secured.Value {
		srv.server.TLSConfig = &tls.Config{
			Rand: rand.Reader,
		}
	}

	srv.logger = slog.New(slog.Default().Handler())

	r := chi.NewRouter()
	r.Use(srv.withLoggingMiddleware)

	r.Get("/ping", srv.pingHandler())
	r.Route("/", func(r chi.Router) {
		r.Use(srv.withEncodingMiddleware)
		r.Use(srv.withAuthorization)

		r.Get("/{id}", srv.newAuditedHandler("follow", srv.unMinifyURLHandler()))
		r.Post("/", srv.newAuditedHandler("shorten", srv.authorized(srv.minifyURLHandler())))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", srv.newAuditedHandler("shorten", srv.authorized(srv.minifyURLJSONHandler())))
			r.Post("/shorten/batch", srv.authorized(srv.minifyURLsHandler()))
			r.Get("/user/urls", srv.authorized(srv.userURLsHandler()))
			r.Delete("/user/urls", srv.authorized(srv.deleteUserURLsHandler()))
		})
	})
	r.Get("/debug/pprof/profile", pprof.Profile)
	r.Method("GET", "/debug/pprof/heap", pprof.Handler("heap"))

	if cfg.AuditFile.Value.enabled {
		srv.auditors = append(srv.auditors, cfg.AuditFile.Value)
	}

	if cfg.AuditURL.Value.enabled {
		srv.auditors = append(srv.auditors, cfg.AuditURL.Value)
	}

	srv.server.Handler = r

	return srv
}

// ListenAndServe starts underlying [http.Server]
func (s *Server) ListenAndServe() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, auditor := range s.auditors {
		for _, auditSubject := range s.auditSubjects {
			go auditor.subscribe(ctx, auditSubject)
			s.logger.Info(fmt.Sprintf("subscribed auditor %+v to subject %+v", auditor, auditSubject))
		}
	}

	s.server.Addr = s.Address.String()
	slog.Info(fmt.Sprintf("Minifier is listening on address: %s\n", s.server.Addr))

	if *s.Secured.Value {
		// * TODO: add certFile, keyFile?
		return s.server.ListenAndServeTLS("", "")
	}

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

type handlerWithUserID func(userID string, res http.ResponseWriter, req *http.Request)

func (s *Server) minifyURLHandler() handlerWithUserID {
	return func(userID string, res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var err error

		if userID == "" {
			err = errors.New("userID is empty")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}

		err = validateContentType("text/plain", req.Header)
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

		s.logger.Info(fmt.Sprintf("Original URL: %s", url))

		minifiedURL, err := s.Service.MinifyURL(ctx, userID, url.String())
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				statusCode = http.StatusInternalServerError
				responseWithError(res, err, statusCode)
				s.logger.Error(err.Error())
				return
			}
			statusCode = http.StatusConflict
		}

		s.logger.Info(fmt.Sprintf("minifiedURL: %s", minifiedURL))

		ctx = context.WithValue(ctx, contextKeyOriginalURL, url.String())
		*req = *req.WithContext(ctx)

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
		ctx := req.Context()
		url, isDeleted, err := s.Service.UnMinifyUserURL(ctx, id)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		if isDeleted {
			res.WriteHeader(http.StatusGone)
			return
		}

		ctx = context.WithValue(ctx, contextKeyOriginalURL, url)
		*req = *req.WithContext(ctx)

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

func (s *Server) minifyURLJSONHandler() handlerWithUserID {
	return func(userID string, res http.ResponseWriter, req *http.Request) {
		var err error

		if userID == "" {
			err = errors.New("userID is empty")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}

		if err = validateContentType("application/json", req.Header); err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}

		// decode JSON request
		var reqBody minifyURLRequest
		d := json.NewDecoder(req.Body)
		if err = d.Decode(&reqBody); err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}
		defer req.Body.Close()

		// handle request
		ctx := req.Context()
		minifiedURL, err := s.Service.MinifyURL(ctx, userID, reqBody.URL)
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				statusCode = http.StatusInternalServerError
				responseWithError(res, err, statusCode)
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

		ctx = context.WithValue(ctx, contextKeyOriginalURL, reqBody.URL)
		*req = *req.WithContext(ctx)

		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		res.WriteHeader(statusCode)
		res.Write([]byte(jsonData))
	}
}

func responseWithError(res http.ResponseWriter, err error, statusCode int) {

	if res == nil {
		err = fmt.Errorf("%w: %s", errResponseWithError, errors.New("nil responseWriter"))
		slog.Error(err.Error())
		return
	}
	res.Header().Set("Content-Type", "text/plain")

	if err == nil {
		// replace the err and HTTP status code and still write the response
		err = fmt.Errorf("%w: %s", errResponseWithError, errors.New("nil err"))
		statusCode = http.StatusInternalServerError
		slog.Error(err.Error())
		http.Error(res, err.Error(), statusCode)
		return
	}

	http.Error(res, err.Error(), statusCode)
}

var errResponseWithError = errors.New("failed to response with error")

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

func (s *Server) minifyURLsHandler() handlerWithUserID {
	return func(userID string, res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var err error

		if userID == "" {
			err = errors.New("userID is empty")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}

		var reqBody minifyURLsRequest
		err = json.NewDecoder(req.Body).Decode(&reqBody)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			return
		}

		// handle request
		var originalURLs []map[string]string
		for _, v := range reqBody {
			originalURLs = append(originalURLs, v.toMap())
		}

		ctx := req.Context()
		minifiedURLs, err := s.MinifyURLs(ctx, userID, originalURLs)
		statusCode := http.StatusCreated
		if err != nil {
			if !errors.Is(err, minifier.ErrMinifiedAlready) {
				statusCode = http.StatusInternalServerError
				responseWithError(res, err, statusCode)
				s.logger.Error(err.Error())
				return
			}
			statusCode = http.StatusConflict
		}

		// handle response
		var resBody minifyURLsResponse
		for _, v := range minifiedURLs {
			var mURL minifyURLsResponseItem
			mURL.fromMap(v)
			resBody = append(resBody, mURL)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(statusCode)
		if err = json.NewEncoder(res).Encode(resBody); err != nil {
			s.logger.Error(err.Error())
		}
	}
}

type (
	minifyURLsRequest  []minifyURLsRequestItem
	minifyURLsResponse []minifyURLsResponseItem
)

type minifyURLsRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

func (m minifyURLsRequestItem) toMap() map[string]string {
	return map[string]string{m.CorrelationID: m.OriginalURL}
}

func (m *minifyURLsResponseItem) fromMap(ma map[string]string) {
	for i, v := range ma {
		m.CorrelationID = i
		m.ShortURL = v
	}
}

type minifyURLsResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s *Server) userURLsHandler() handlerWithUserID {
	return func(userID string, res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var err error

		if userID == "" {
			err = errors.New("userID is empty")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}

		ctx := req.Context()

		userURLs, err := s.UserURLs(ctx, userID)
		if err != nil {
			responseWithError(res, err, http.StatusInternalServerError)
			s.logger.Error(err.Error())
			return
		}

		statusCode := http.StatusNoContent
		var r userURLsHandlerResponse
		for _, v := range userURLs {
			r = append(r,
				userURLsResponseItem{
					ShortURL:    v.MinifiedURL.String(),
					OriginalURL: v.OriginalURL.String()},
			)
		}

		if len(r) > 0 {
			statusCode = http.StatusOK
		}

		s.responseWithJSON(res, r, statusCode)

	}
}

func (s *Server) deleteUserURLsHandler() handlerWithUserID {
	return func(userID string, res http.ResponseWriter, req *http.Request) {
		var err error

		req = req.WithContext(req.Context())

		if userID == "" {
			err = errors.New("userID is empty")
			responseWithError(res, err, http.StatusUnauthorized)
			s.logger.Error(err.Error())
			return
		}
		defer req.Body.Close()

		err = validateContentType("application/json", req.Header)
		if err != nil {
			responseWithError(res, err, http.StatusBadRequest)
			s.logger.Error(err.Error())
			return
		}
		var minifiedIDs []string
		d := json.NewDecoder(req.Body)
		err = d.Decode(&minifiedIDs)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				responseWithError(res, err, http.StatusBadRequest)
				s.logger.Error(err.Error())
				return
			}
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err := s.Service.DeleteUserURLs(ctx, userID, minifiedIDs)
			if err != nil {
				s.logger.Error(err.Error())
			}
		}()

		res.WriteHeader(http.StatusAccepted)
	}
}

type (
	userURLsHandlerResponse []userURLsResponseItem

	userURLsResponseItem struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

func (s *Server) responseWithJSON(res http.ResponseWriter, payload any, statusCode int) {
	var (
		jsonData []byte
		err      error
	)

	if jsonData, err = json.Marshal(payload); err != nil {
		err = fmt.Errorf("%w: %s", errResponseWithJSON, err)
		statusCode = http.StatusInternalServerError
		responseWithError(res, err, statusCode)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	res.Write(jsonData)

}

var errResponseWithJSON = errors.New("failed to respond with JSON")
