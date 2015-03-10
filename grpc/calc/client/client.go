package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/yml/sandbox/grpc/calc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "127.0.0.1:9999", "The server addr")
)

func main() {
	fmt.Println("Starting the rpc client")
	flag.Parse()
	conn, err := grpc.Dial(*addr)
	if err != nil {
		log.Fatalf("Failed to connect to the grpc server, err: %v", err)
	}
	defer conn.Close()
	client := proto.NewCalcClient(conn)
	resp, err := client.Add(context.Background(), &proto.Request{X: 2, Y: 3})
	if err != nil {
		log.Fatalf("Failed to add with err : %v", err)
	}
	fmt.Printf("Result of the addition %v/n", resp.Z)
}
