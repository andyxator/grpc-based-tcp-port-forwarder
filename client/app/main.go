package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/andyxator/grpc-based-tcp-port-forwarder/client"

	"google.golang.org/grpc"
)

func main() {
	exitCh := make(chan struct{})

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		//grpc.WithBlock(),
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", 7777), opts...)
	if err != nil {
		log.Print(err)
		return
	}

	log.Print("Forwarder started")

	ctx, _ := context.WithCancel(context.Background())
	go func() {
		defer func() {
			exitCh <- struct{}{}
		}()

		err := listenAndForward(
			ctx,
			clientConn,
			"localhost", 8080,
			"www.detroitindustrial.org", 8008,
		)

		if err != nil {
			log.Print(err)
		}
	}()

	//time.Sleep(time.Second)
	//cancel()

	<-exitCh
}

func listenAndForward(ctx context.Context, clientConn *grpc.ClientConn, lhost string, lport int32, rhost string, rport int32) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	laddr := fmt.Sprintf("%s:%d", lhost, lport)
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalf(err.Error())
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	log.Printf("Forwarder start listening on '%s'", listener.Addr())
	defer func() {
		log.Printf("Forwarder stop listening on '%s'", listener.Addr())
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			log.Printf("Client '%v' connected", conn.RemoteAddr())

			defer func() {
				log.Printf("Client '%v' disconnected", conn.RemoteAddr())
				conn.Close()
			}()

			client, err := client.NewClient(clientConn)
			if err != nil {
				return
			}

			err = client.Forward(ctx, conn, rhost, rport)
			if err != nil {
				return
			}
		}()
	}
}
