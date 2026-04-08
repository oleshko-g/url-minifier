package grpc

import (
	"context"

	"github.com/oleshko-g/url-minifier/internal/transport/config"
	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	*grpc.Server
}

type server struct {
	minifier_v1.UnimplementedMinifierServiceServer
}

func NewServer(cfg *config.Config) *Server {
	srv := grpc.NewServer()
	minifier_v1.RegisterMinifierServiceServer(srv, &server{})
	return &Server{}
}

func (s *server) MinifyURL(ctx context.Context, req *minifier_v1.MinifyURLRequest) (*minifier_v1.MinifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.MinifyURL(ctx, req)
}

func (s *server) UnminifyURL(ctx context.Context, req *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.UnminifyURL(ctx, req)
}

func (s *server) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	return s.UnimplementedMinifierServiceServer.ListUserURLs(ctx, req)
}
