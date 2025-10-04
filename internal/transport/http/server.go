package http

import (
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
		err := validateContentType(req.Header, "text/plain")
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

func respondBadRequest(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	log.Printf("error: %s", err)
}

func validateContentType(h http.Header, value string) error {
	mediaType, _, errParseMediaType := mime.ParseMediaType(h.Get("Content-Type"))
	if errParseMediaType != nil {
		return errParseMediaType
	}

	if mediaType != value {
		return errors.New("error: Content-Type isn't " + value)
	}

	return nil
}
