package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

func fib(n int64) int64 {
	if n <= 2 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func fibServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(fmt.Errorf("An error occured while listening to: %s -- %s", addr, err))
	}
	for {
		conn, err := ln.Accept()
		log.Println("connection", addr)
		if err != nil {
			log.Println(err)
			continue
		}
		go fibHandler(conn)
	}
}

func fibHandler(conn net.Conn) {
	buf := make([]byte, 100)
	var req int
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			conn.Close()
			break
		}
		req, err = strconv.Atoi(string(bytes.Trim(buf[0:n], "\n")))
		if err != nil {
			log.Println("The request must be a number", err)
			continue
		}
		resp := strconv.FormatInt(fib(int64(req)), 10)
		_, err = conn.Write([]byte(resp + "\n"))
		if err != nil {
			fmt.Println("Error while writing to the socket")
		}
	}
	log.Println("closed")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("You must provide an addr (ie: 127.0.0.1:25000 )")
		return
	}
	fibServer(args[0])
}
