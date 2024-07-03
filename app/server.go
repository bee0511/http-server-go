package main

import (
    "fmt"
    "net"
    "os"
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
	response := "HTTP/1.1 200 OK\r\n" +
	"Content-Length: 0\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" // End of headers.
    _, err := conn.Write([]byte(response))
    if err != nil {
        fmt.Println("Error writing response:", err)
    }
}
