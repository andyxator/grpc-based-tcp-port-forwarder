package server

import (
	"fmt"
	"log"
	"net"

	pb "github.com/andyxator/grpc-based-tcp-port-forwarder/proto"
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func (s *server) Forward(stream pb.Proxy_ForwardServer) error {

	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", msg.Addr.Host, msg.Addr.Port))
	if err != nil {
		log.Print("could not connect to target", err)
		return err
	}
	defer target.Close()

	go func() {

		for {

			data := make([]byte, 1024)
			n, err := target.Read(data)
			if err != nil {
				return
			}

			chunk := &pb.BytesChunk{
				Data: data[:n],
			}

			err = stream.Send(chunk)
			if err != nil {
				return
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		_, err = target.Write(msg.Chunk.Data)
		if err != nil {
			return err
		}
	}

	return nil
}
