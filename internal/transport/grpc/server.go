package grpc

import (
	"context"

	minifier_v1 "github.com/oleshko-g/url-minifier/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	minifier_v1.MinifierServiceServer
}

func New() *Server {
	return &Server{}
}

func (s *Server) MinifyURL(context.Context, *minifier_v1.MinifyURLRequest) (*minifier_v1.MinifyURLResponse, error) {
	return nil, nil
}

func (s *Server) UnminifyURL(context.Context, *minifier_v1.UnminifyURLRequest) (*minifier_v1.UnminifyURLResponse, error) {
	return nil, nil
}

func (s *Server) ListUserURLs(context.Context, *emptypb.Empty) (*minifier_v1.UserURLsResponse, error) {
	return nil, nil
}
