package main

import (
    "fmt"
    "net"
    "os"
	"strings"
)

func main() {
    l, err := net.Listen("tcp", "0.0.0.0:4221")
    if err != nil {
        fmt.Println("Failed to bind to port 4221:", err)
        os.Exit(1)
    }
    defer l.Close()
    fmt.Println("Listening on port 4221...")

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
	req := make([]byte, 1024)
	conn.Read(req)
	if strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
