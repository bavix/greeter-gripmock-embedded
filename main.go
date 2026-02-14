// Package main runs TimedGreeter gRPC server.
package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bavix/greeter-gripmock-embedded/helloworld"
	"github.com/bavix/greeter-gripmock-embedded/timed"
)

func NewTimedGreeterServer(greeterClient helloworld.GreeterClient) timed.TimedGreeterServer {
	return &timedGreeterServer{greeterClient: greeterClient}
}

type timedGreeterServer struct {
	timed.UnimplementedTimedGreeterServer

	greeterClient helloworld.GreeterClient
}

func (s *timedGreeterServer) SayHello(ctx context.Context, req *timed.HelloRequest) (*timed.HelloReply, error) {
	start := time.Now()

	hwReply, err := s.greeterClient.SayHello(ctx, &helloworld.HelloRequest{Name: req.GetName()})
	if err != nil {
		return nil, err
	}

	return &timed.HelloReply{
		Message:    hwReply.GetMessage(),
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

func main() {
	greeterAddr := "localhost:50051"
	if addr := os.Getenv("GREETER_ADDR"); addr != "" {
		greeterAddr = addr
	}

	conn, err := grpc.NewClient(greeterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial Greeter: %v", err)
	}

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:50052")
	if err != nil {
		_ = conn.Close()

		log.Fatalf("listen: %v", err)
	}

	defer func() { _ = conn.Close() }()
	defer func() { _ = lis.Close() }()

	srv := grpc.NewServer()
	timed.RegisterTimedGreeterServer(srv, NewTimedGreeterServer(helloworld.NewGreeterClient(conn)))

	go func() {
		log.Printf("TimedGreeter listening on %s", lis.Addr())
		_ = srv.Serve(lis)
	}()

	<-waitShutdown()
	srv.GracefulStop()
}

func waitShutdown() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	return quit
}
