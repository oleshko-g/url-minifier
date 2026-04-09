package grpc

import (
	"context"
	"net"

	"github.com/oleshko-g/url-minifier/internal/service/minifier"
	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	*grpc.Server
	*Config
	*service
}

type service struct {
	minifier_v1.UnimplementedMinifierServiceServer
	srvc minifier.Service
}

func NewServer(cfg *Config, srvc minifier.Service) *Server {
	srv := grpc.NewServer()
	minifier_v1.RegisterMinifierServiceServer(srv, &service{
		srvc: srvc,
	})

	return &Server{
		Server:  srv,
		Config:  cfg,
		service: &service{srvc: srvc},
	}
}

// ListenAndServe creates a listener on the configured address and serves incoming gRPC requests.
func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.Address.Value.String())
	if err != nil {
		return err
	}

	return s.Server.Serve(lis)
}

func (s *Server) GracefulShutdown() error {
	s.Server.GracefulStop()
	return s.service.srvc.Close()
}

func (s *service) MinifyURL(ctx context.Context, req *minifier_v1.MinifyURLRequest) (*minifier_v1.MinifyURLResponse, error) {
	s.srvc.MinifyURL(ctx, "", req.GetUrl())
	return s.UnimplementedMinifierServiceServer.MinifyURL(ctx, req)
}

func (s *service) UnminifyURL(ctx context.Context, req *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.UnminifyURL(ctx, req)
}

func (s *service) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	return s.UnimplementedMinifierServiceServer.ListUserURLs(ctx, req)
}
