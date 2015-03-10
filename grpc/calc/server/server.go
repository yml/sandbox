package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yml/sandbox/grpc/calc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9999, "The server port")
)

type calcServer struct{}

func (c *calcServer) Add(ctx context.Context, r *proto.Request) (*proto.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	return &proto.Response{Z: r.X + r.Y}, nil
}

func main() {
	fmt.Println("Starting the server")
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to bind to port %d, err %v", *port, err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterCalcServer(grpcServer, &calcServer{})
	grpcServer.Serve(l)

}
