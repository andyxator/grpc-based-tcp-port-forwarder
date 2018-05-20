package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/andyxator/grpc-based-tcp-port-forwarder/proto"
	"github.com/andyxator/grpc-based-tcp-port-forwarder/server"

	"google.golang.org/grpc"
)

func main() {
	exitCh := make(chan struct{})

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterProxyServer(grpcServer, server.NewServer())

	go func() {
		defer close(exitCh)
		grpcServer.Serve(lis)
	}()

	<-exitCh
}
