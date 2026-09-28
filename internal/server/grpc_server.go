package server

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"gmhelper.solution-hub/internal/handler"
	pb "gmhelper.solution-hub/proto"
)

func RunGRPC(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	grpcServer := grpc.NewServer()

	solveHandler := handler.NewSolveHandler()
	pb.RegisterSolutionHubServer(grpcServer, solveHandler)

	log.Println("gRPC server started on port", port)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
