package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/oleshko-g/url-minifier/internal/service"
	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	"github.com/oleshko-g/url-minifier/internal/transform"
	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	*grpc.Server
	*Config
	*implemented
}

type implemented struct {
	minifier_v1.UnimplementedMinifierServiceServer
	service.Minifier
}

func NewServer(cfg *Config, minifier service.Minifier) *Server {
	s := &Server{
		Config:      cfg,
		implemented: &implemented{Minifier: minifier},
	}
	s.Server = grpc.NewServer(s.authOption())

	minifier_v1.RegisterMinifierServiceServer(s.Server, s.implemented)

	return s
}

// ListenAndServe creates a listener on the configured address and serves incoming gRPC requests.
func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.Address.Value.String())
	if err != nil {
		return err
	}

	return s.Server.Serve(lis)
}

func (s *Server) GracefulStop() error {
	s.Server.GracefulStop()
	return s.implemented.Minifier.Close()
}

func (s *implemented) MinifyURL(ctx context.Context, req *minifier_v1.MinifyURLRequest) (*minifier_v1.MinifyURLResponse, error) {
	userID := metadata.ValueFromIncomingContext(ctx, "userID")
	if len(userID) != 1 {
		return nil, ErrUnauthenticated
	}

	minifiedURL, err := s.Minifier.MinifyURL(ctx, userID[0], req.GetUrl())
	if err != nil {
		if !errors.Is(err, minifier.ErrMinifiedAlready) {
			return nil, ErrInternalServerError
		}

		return &minifier_v1.MinifyURLResponse{Result: minifiedURL}, err
	}

	return &minifier_v1.MinifyURLResponse{Result: minifiedURL}, nil
}

func (s *implemented) UnminifyURL(ctx context.Context, req *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	userID := metadata.ValueFromIncomingContext(ctx, "userID")
	if len(userID) != 1 {
		return nil, ErrUnauthenticated
	}

	originalURL, isDeleted, err := s.Minifier.UnMinifyUserURL(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternalServerError, err)
	}

	if isDeleted {
		return nil, ErrNotFound
	}

	return &minifier_v1.UnminifyURLResponse{
		Result: originalURL,
	}, nil
}

func (s *implemented) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	userID := metadata.ValueFromIncomingContext(ctx, "userID")
	if len(userID) != 1 {
		return nil, ErrUnauthenticated
	}
	URLs, err := s.Minifier.UserURLs(ctx, userID[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternalServerError, err)
	}

	URLData := transform.SliceToSlice(URLs, func(mURL minifier.URL) *minifier_v1.URLData {
		return &minifier_v1.URLData{
			OriginalUrl: mURL.OriginalURL.String(),
			ShortUrl:    mURL.MinifiedURL.String(),
		}
	})

	return &minifier_v1.UserURLsResponse{
		Url: URLData,
	}, nil
}
