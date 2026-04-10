package grpc

import (
	"context"
	"net"

	"github.com/oleshko-g/url-minifier/internal/service"
	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/grpc"
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
	s.Minifier.MinifyURL(ctx, "", req.GetUrl())
	return s.UnimplementedMinifierServiceServer.MinifyURL(ctx, req)
}

func (s *implemented) UnminifyURL(ctx context.Context, req *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.UnminifyURL(ctx, req)
}

func (s *implemented) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	return s.UnimplementedMinifierServiceServer.ListUserURLs(ctx, req)
}
