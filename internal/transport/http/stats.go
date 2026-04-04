package http

import "net/http"

func (s *Server) statsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// * TODO: check request header X-Trusted-Subnet
		// * TODO: go getNumberOfMinifiedURLs
		// * TODO: go getNumberOfUsers
		defer r.Body.Close()
		w.WriteHeader(http.StatusNotImplemented)
	}
}
