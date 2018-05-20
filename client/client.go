package client

import (
	"context"
	"log"
	"net"

	pb "github.com/andyxator/grpc-based-tcp-port-forwarder/proto"

	"google.golang.org/grpc"
)

type client struct {
	grpcClient pb.ProxyClient
}

func NewClient(conn *grpc.ClientConn) (*client, error) {
	return &client{
		grpcClient: pb.NewProxyClient(conn),
	}, nil
}

func (s *client) Forward(ctx context.Context, conn net.Conn, host string, port int32) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := s.grpcClient.Forward(ctx)
	if err != nil {
		return err
	}

	msg := &pb.ProxyMessage{
		Addr: &pb.TargetAddr{
			Host: host,
			Port: port,
		},
	}

	err = stream.Send(msg)
	if err != nil {
		return err
	}

	go func() {
		defer cancel()
		for {
			data := make([]byte, 1024)
			n, err := conn.Read(data)
			if err != nil {
				return
			}

			msg := &pb.ProxyMessage{
				Chunk: &pb.BytesChunk{
					Data: data[:n],
				},
			}

			err = stream.Send(msg)
			if err != nil {
				return
			}
		}
	}()

	connected := false
	for {
		chunk, err := stream.Recv()
		if err != nil {
			return err
		}

		_, err = conn.Write(chunk.Data)
		if err != nil {
			return err
		}

		if !connected {
			connected = true
			log.Printf("Connection '%s' <-> '%s:%d' established", conn.RemoteAddr(), msg.Addr.Host, msg.Addr.Port)
		}
	}

	return nil
}
