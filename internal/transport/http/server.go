package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Server struct {
	server *http.Server
	service
	Config
	zerolog.Logger
}

type service interface {
	MinifyURL(url string) (minifiedURL string, err error)
	UnMinifyURL(id string) (url string, err error)
}

func NewServer(s service) *Server {
	srv := &Server{
		service: s,
		server:  &http.Server{},
	}

	r := chi.NewRouter()
	r.Post("/",
		srv.withLoggingMiddleware(
			srv.minifyURLHandler()))
	r.Get("/{id}",
		srv.withLoggingMiddleware(
			srv.unMinifyURLHandler()))
	srv.server.Handler = r

	r.Post("/api/shorten",
		srv.withLoggingMiddleware(
			srv.minifyURLJSONHandler()))
	srv.server.Handler = r

	zl := zerolog.New(os.Stderr).With().Timestamp().Logger()
	srv.Logger = zl

	return srv
}

func (s *Server) ListenAndServe() error {
	s.server.Addr = s.Address().String()
	log.Printf("Minifier is listening on address: %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s Server) minifyURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := validateContentType("text/plain", req.Header)
		if err != nil {
			respondBadRequest(res, err)
			return
		}

		data, err := io.ReadAll(req.Body)
		if err != nil {
			respondBadRequest(res, err)
			return
		}

		url, err := url.Parse(string(data))
		if err != nil {
			respondBadRequest(res, err)
			return
		}

		log.Printf("Original URL: %s", url)

		minifiedURL, err := s.service.MinifyURL(url.String())
		if err != nil {
			respondBadRequest(res, err)
			return
		}
		log.Printf("minifiedURL: %s", minifiedURL)
		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Content-Length", strconv.Itoa(len(minifiedURL)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(minifiedURL))
		log.Printf("%+v\n", res)
	}
}

func (s Server) unMinifyURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")
		if id == "" {
			respondBadRequest(res, errors.New("empty id"))
			return
		}
		url, err := s.service.UnMinifyURL(id)
		if err != nil {
			respondBadRequest(res, errors.New("URL is not found"))
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

func (s Server) minifyURLJSONHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if err := validateContentType("application/json", req.Header); err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}

		// decode JSON request
		var reqBody minifyURLRequest
		d := json.NewDecoder(req.Body)
		if err := d.Decode(&reqBody); err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}
		defer req.Body.Close()

		// handle request
		minifiedURL, err := s.service.MinifyURL(reqBody.URL)
		if err != nil {
			respondInternalServerError(res, err)
			s.Logger.Err(err).Msg("")
			return
		}

		// encode response
		resBody := minifyURLResponse{
			Result: minifiedURL,
		}
		jsonData, err := json.Marshal(&resBody)
		if err != nil {
			respondInternalServerError(res, err)
			s.Logger.Err(err).Msg("")
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(jsonData))
	}
}

func respondBadRequest(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	log.Printf("error: %s", err)
}

func respondInternalServerError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusInternalServerError)
	log.Printf("error: %s", err)
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
