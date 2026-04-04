package http

import (
	"net"
	"net/http"
)

func (s *Server) statsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// * TODO: check trusted subnet
		defer r.Body.Close()

		realIPString := r.Header.Get("X-Real-IP")
		if realIPString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		IPAdress, IPsubnet, err := net.ParseCIDR(realIPString)
		if err != nil {
			s.logger.Error(err.Error())
			responseWithError(w, err, http.StatusInternalServerError)
			return
		}

		_, _ = IPAdress, IPsubnet

		// * TODO: check request header X-Trusted-Subnet

		// * TODO: go getNumberOfMinifiedURLs

		// * TODO: go getNumberOfUsers

		w.WriteHeader(http.StatusNotImplemented)
	}
}
