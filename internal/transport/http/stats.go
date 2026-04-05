package http

import (
	"context"
	"net"
	"net/http"
	"sync"
)

// subnet represents a trusted IP subnet in CIDR notation.
type subnet struct {
	net.IP
	*net.IPNet
}

// Set parses the given string as a CIDR notation IP subnet and sets it as the trusted subnet.
// If the string is not a valid CIDR notation, it returns an error.
func (t *subnet) Set(s string) error {

	ip, network, err := net.ParseCIDR(s)
	if err != nil {
		return err
	}

	t.IP = ip
	t.IPNet = network

	return nil
}

// String returns a string representation of the trusted IP subnet in CIDR notation.
// If the subnet is not set, it returns an empty string.
func (t *subnet) String() string {
	if t.IPNet == nil {
		return ""
	}

	return t.IPNet.String()
}

func (s *Server) statsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// * TODO: check trusted subnet
		defer r.Body.Close()

		if s.Config.TrustedIPSubnet.Value == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		realIPString := r.Header.Get("X-Real-IP")
		if realIPString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ip, _, err := net.ParseCIDR(realIPString)
		if err != nil {
			s.logger.Error(err.Error())
			responseWithError(w, err, http.StatusInternalServerError)
			return
		}

		if !s.Config.TrustedIPSubnet.Value.IPNet.Contains(ip) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithCancelCause(r.Context())
		defer cancel(nil)

		var res StatsResponse

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			users, err := s.TotalUsers(ctx)
			if err != nil {
				cancel(err)
				return
			}
			res.Users = users
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			URLs, err := s.TotalURLs(ctx)
			if err != nil {
				cancel(err)
				return
			}
			res.URLs = URLs
		}()

		wg.Wait()

		s.responseWithJSON(w, res, http.StatusOK)

	}
}

type StatsResponse struct {
	Users int `json:"users"`
	URLs  int `json:"urls"`
}
