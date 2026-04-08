package grpc

import (
	"context"

	"github.com/oleshko-g/url-minifier/internal/transport/config"
	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	minifier_v1.UnimplementedMinifierServiceServer
}

func New(cfg *config.Config) *Server {
	return &Server{}
}

func (s *Server) MinifyURL(ctx context.Context, req *minifier_v1.MinifyURLRequest) (*minifier_v1.MinifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.MinifyURL(ctx, req)
}

func (s *Server) UnminifyURL(ctx context.Context, req *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	return s.UnimplementedMinifierServiceServer.UnminifyURL(ctx, req)
}

func (s *Server) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	return s.UnimplementedMinifierServiceServer.ListUserURLs(ctx, req)
}
