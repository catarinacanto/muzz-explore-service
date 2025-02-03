package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"muzz-explore-service/internal/service"
	pb "muzz-explore-service/pkg/pb/proto"
)

type Server struct {
	grpcServer *grpc.Server
}

func NewGRPCServer(service *service.ExploreService) *Server {
	server := grpc.NewServer()
	pb.RegisterExploreServiceServer(server, service)

	// Enable reflection
	reflection.Register(server)

	return &Server{
		grpcServer: server,
	}
}

func (s *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
