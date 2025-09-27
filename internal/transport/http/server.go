package http

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
)

type Server struct {
	Service
	*http.Server
	Config
}

type Service interface {
	MinifyURL(url string) (minifiedURL string, err error)
	UnMinifyURL(id string) (url string, err error)
}

func NewServer(s Service) *Server {
	srv := &Server{
		Service: s,
		Server:  &http.Server{},
	}
	r := chi.NewRouter()
	r.Post("/", srv.minifyURLHandler())
	r.Get("/{id}", srv.unMinifyURLHandler())
	srv.Server.Handler = r
	return srv
}

func (s *Server) ListenAndServe() error {
	s.Server.Addr = s.Address().String()
	log.Printf("Minifier is listening on address: %s\n", s.Server.Addr)
	return s.Server.ListenAndServe()
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

		minifiedURL, err := s.Service.MinifyURL(url.String())
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
		url, err := s.Service.UnMinifyURL(id)
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
