# concurrency in Python vs GO

I have been lucky enough to attend Pycon in Montreal in the past few days among all the talks I attended one has blown my mind away: [Python concurrency from the Ground Up: LIVE!](http://us.pycon.org/2015/schedule/presentation/374/) by David Beazley. The video is available on [youtube](https://www.youtube.com/watch?v=MCs5OvhV9S4)

The gist of the talk is that going from synchronous program in python to a concurrent program requires a significant amount of leg work. The talk take a simple socket program that calculate **fibonacci** sum synchronously and makes it concurrents. The talk compare and contrast various approach: Threads, Multi processes, corountines.

My take away was that there is multiple way of doing it in python but none of them are great at taking advantage of multi cores.
When I went through the process of typing the code used in his demo I decided for the fun of it to port to GO to compare and contrast.

The first surprises for me was how similar the synchronous version looks like in both language.

```
# synchronous.py
from socket import *

def fib(n):
    if n <= 2:
        return 1
    else:
        return fib(n-1) + fib(n-2)

def fib_server(address):
    sock = socket(AF_INET, SOCK_STREAM)
    sock.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
    sock.bind(address)
    sock.listen(5)
    while True:
        conn, addr = sock.accept() # blocking
        print("connection", addr)
        fib_handler(conn)

def fib_handler(conn):
    while True:
        req = conn.recv(100)  # blocking
        if not req:
            break
        n = int(req)
        result = fib(n)
        resp = str(result).encode('ascii') + b'\n'
        conn.send(resp)  # blocking
    print('closed')

if __name__ == "__main__":
    fib_server(('', 25000))
```

The GO version requires a bit more typing and type ceremonies.

```
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
		fibHandler(conn)
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
		reqStr := string(bytes.Trim(buf[0:n], "\n"))
		req, err = strconv.Atoi(reqStr)
		if err != nil {
			log.Println("The request must be a number", reqStr, err)
		}
		result := fmt.Sprintf("%v\n", fib(int64(req)))
		_, err = conn.Write([]byte(result))
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
		fmt.Println("You must provide an addr (127.0.0.1:25000)")
		return
	}
	fibServer(args[0])
}
```

The beauty of the GO version is that it only takes 2 letters to move from the synchronous to a concurrent version `GO`.

The python equivalent is way harder to pull off.

```
from socket import *
from collections import deque
from select import select

def fib(n):
    if n <= 2:
        return 1
    else:
        return fib(n-1) + fib(n-2)

def fib_server(address):
    sock = socket(AF_INET, SOCK_STREAM)
    sock.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
    sock.bind(address)
    sock.listen(5)
    while True:
        yield 'recv', sock
        conn, addr = sock.accept() # blocking
        print("connection", addr)
        tasks.append(fib_handler(conn))

def fib_handler(conn):
    while True:
        yield 'recv', conn
        req = conn.recv(100)  # blocking
        if not req:
            break
        n = int(req)
        result = fib(n)
        resp = str(result).encode('ascii') + b'\n'
        yield 'send', conn
        conn.send(resp)  # blocking
    print('closed')

tasks = deque()
recv_wait = {}
send_wait = {}

def run():
    while any([tasks, recv_wait, send_wait]):
        while not tasks:
            # no active task to run wait for IO
            can_recv, can_send, _ = select(recv_wait, send_wait, [])
            for s in can_recv:
                tasks.append(recv_wait.pop(s))
            for s in can_send:
                tasks.append(send_wait.pop(s))
        task = tasks.popleft()
        try:
            why, what = next(task)
            if why == 'recv':
                recv_wait[what] = task
            elif why == 'send':
                send_wait[what] = task
            else:
                raise RuntimeError("We don't know what to do with :", why)
        except StopIteration:
            print('task done')

if __name__ == "__main__":
    tasks.append(fib_server(('localhost', 25000)))
    run()
```

The interesting part is that even with all this works the python version can't take advantage of all the cores. Where the GO equivalent is controlled by an env variable called `GOMAXPROCS` that determined how many cores you want to allocate to your programs.
