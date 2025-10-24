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
	*Config
	zerolog.Logger
}

type service interface {
	MinifyURL(url string) (minifiedURL string, err error)
	UnMinifyURL(id string) (url string, err error)
}

func NewServer(s service, cp *Config) *Server {
	srv := &Server{
		service: s,
		server:  &http.Server{},
		Config:  cp,
	}

	srv.Config.canDecompress = map[coding]struct{}{codingGZIP: {}}
	srv.Config.canCompress = []coding{codingGZIP, codingIdentity}

	r := chi.NewRouter()
	r.Use()
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

	zl := zerolog.New(os.Stderr).With().Timestamp().Logger()
	srv.Logger = zl

	return srv
}

func (s *Server) ListenAndServe() error {
	s.server.Addr = s.Address().String()
	log.Printf("Minifier is listening on address: %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// canDecompress reports if the server can decompress the given compression format
func (s *Server) canDecompress(compression coding) bool {
	_, ok := s.Config.canDecompress[compression]
	return ok
}

func (s *Server) chooseCompression(parsedAcceptCodings map[coding]qualityValue) (coding, error) {
	compression := parsedCoding{
		coding:       s.Config.canCompress[0],
		qualityValue: 1.0,
	}

	for _, c := range s.canCompress {
		// check if the client explicitly specified a coding which the server [canCompress]
		if q, ok := parsedAcceptCodings[c]; ok {

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

		// check if the client implicitly specified a coding which the server [canCompresss]
		if q, ok := parsedAcceptCodings[codingWildcard]; ok {
			if compression.qualityValue != q {
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
		return "", errors.New("the client has forbidden every coding which the server can compress the response with")
	}

	return compression.coding, nil
}

func (s Server) minifyURLHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := validateContentType("text/plain", req.Header)
		if err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}

		data, err := io.ReadAll(req.Body)
		if err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}

		url, err := url.Parse(string(data))
		if err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}

		log.Printf("Original URL: %s", url)

		minifiedURL, err := s.service.MinifyURL(url.String())
		if err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
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
		var err error
		id := req.PathValue("id")
		if id == "" {
			err = errors.New("empty id")
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
			return
		}
		url, err := s.service.UnMinifyURL(id)
		if err != nil {
			respondBadRequest(res, err)
			s.Logger.Err(err).Msg("")
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
}

func respondInternalServerError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusInternalServerError)
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
