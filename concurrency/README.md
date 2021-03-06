# concurrency in Python vs GO

At Pycon in Montreal few weeks ago I  attended a talk that blew my mind away and got me thinking: [Python concurrency from the Ground Up: LIVE!](http://us.pycon.org/2015/schedule/presentation/374/) by David Beazley. The video is available on [YouTube](https://www.youtube.com/watch?v=MCs5OvhV9S4).

The gist of the talk is that going from a synchronous to a concurrent program in Python requires a significant amount of leg work. The talk took a simple socket program that calculates the **Fibonacci** sum synchronously and tries to make it concurrent. It compares and contrasts various approachs: threads, multiple processes, and corountines.

My take away was that there are a zillion ways of doing it in Python but none of them are great at taking advantage of multi cores. When I went through the process of typing the code used in his demo I decided for the fun of it to port it to Go.

The first surprises for me was how similar the synchronous version is in both languages. The code and the micro benchmarks that follow should be taken with a grain of salt like always.


## Synchronous

```
@syntax:python
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

The Go version requires a bit more typing and type ceremonies but the structure is very similar.

```
@syntax:go
// synchronous.go
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
		fibHandler(conn)  // prefix by `go` to get concurrent.go
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

### Synchronous Micro-Benchmark 

The benchmark consists of running one instance of `perf2.py` which simulates a client hammering on our micro service.

```
@syntax:python
#perf2.py

from socket import *
import time

from threading import Thread

sock = socket(AF_INET, SOCK_STREAM)
sock.connect(('localhost', 25000))

n= 0

def monitor():
    global n
    while True:
        time.sleep(1)
        print(n, 'reqs/s')
        n = 0

Thread(target=monitor).start()

while True:
    sock.send(b'1')
    resp = sock.recv(100)
    n += 1
```

* **Python** ~17,000 req/s 
* **PyPy**  ~22,000 req/s
* **Go**  ~21,000 req/s

PyPy is faster than go by a small margin but as far as I am concerned I would say that the 3 solutions are within the same order of magnitude.

## Concurrency

The beauty of Go is that it only takes 2 letters to move from a synchronous to a concurrent version. Simply add `go` in front of the function call to `fibHandler(conn)`. Not only is it simple, but, unlike Python, there is one obvious way to do it.

The Python equivalent is way harder to pull off, one could argue that it is probably out of reach for a huge portion of experienced Python developers. David Beazley illustrates very well the phenomenal diversity of approaches that could be taken, all broken to some extent. I am sure some other candidates comes to your mind: asyncio, Twisted, Tornado, etc.

Below you can see the coroutines version with a zest of ProcessPoolExecutor.

```
@syntax:python
from socket import *
from collections import deque
from concurrent.futures import ProcessPoolExecutor as Pool
from select import select

pool = Pool(4)

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
        future = pool.submit(fib, n)
        yield 'future', future 
        result = future.result()  # blocking
        resp = str(result).encode('ascii') + b'\n'
        yield 'send', conn
        conn.send(resp)  # blocking
    print('closed')

tasks = deque()
recv_wait = {}
send_wait = {}
future_wait = {}

future_notify, future_event = socketpair()

def future_done(future):
    tasks.append(future_wait.pop(future))
    future_notify.send(b'x')

def future_monitor():
    while True:
        yield 'recv', future_event
        future_event.recv(100)

tasks.append(future_monitor())

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
            elif why == 'future':
                future_wait[what] = task
                what.add_done_callback(future_done)
            else:
                raise RuntimeError("We don't know what to do with :", why)
        except StopIteration:
            print('task done')

if __name__ == "__main__":
    tasks.append(fib_server(('localhost', 25000)))
    run()
```

The interesting part is that even with all this work the Python version can't take advantage of all the cores. Where the Go equivalent is controlled by an environment variable called `GOMAXPROCS` that determines how many cores you want to allocate to your programs. The performance characteristics are also different by an order of magnitude:

### Concurrent Micro-Benchmark

This micro-benchmark does not include `PyPy` because some of the features used in `concurrency.py` are not currently supported, specifically the `concurrent` module.

#### `fib(30)`

A single iteration with 30 as the argument to `fib`.

* **Python** 231ms
* **Go** 5ms

#### Requests per second

3 clients running `perf2.py`

* **Python** 275 req/s per `perf2.py` instance -- `concurrency.py` takes 188MB of RAM
* **Go** (`GOMAXPROCS=3`): 12500 req/s  per `perf2.py` instance -- `concurrency.go` takes 120MB of RAM

Go is significantly faster than Python; this is fine and expected. What I find more disturbing is how much easier it is to morph a synchronous program into its concurrent equivalent. In addition the resulting piece of Go code is also more readable and easier to reason about. Not all problems require a concurrent solution but for the ones that do Go has a lot to offer.

