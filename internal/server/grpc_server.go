package server

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"gmhelper.solution-hub/internal/handler"
	pb "gmhelper.solution-hub/proto"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func NewServer(port string, solveHandler *handler.SolveHandler) *Server {
	grpcServer := grpc.NewServer()
	pb.RegisterSolutionHubServer(grpcServer, solveHandler)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	log.Printf("[gRPC Server] Starting SolutionHub gRPC server on port %s", s.port)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
