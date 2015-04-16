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
