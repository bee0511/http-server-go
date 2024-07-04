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

    // Buffer to read the request
    req := make([]byte, 1024)
    n, err := conn.Read(req)
    if err != nil {
        fmt.Println("Error reading request:", err)
        return
    }

	if strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}

    // Convert the request to a string and trim any extra spaces
    requestLine := string(req[:n])
    requestLine = strings.TrimSpace(requestLine)

    // Check if the request is a GET /echo/{str} HTTP/1.1
    if strings.HasPrefix(requestLine, "GET /echo/") {
        // Extract the echoed string
        pathParts := strings.Split(requestLine, " ")
        if len(pathParts) > 1 {
            path := pathParts[1]
            if strings.HasPrefix(path, "/echo/") {
                echoStr := strings.TrimPrefix(path, "/echo/")

                // Respond with the echoed string
                response := "HTTP/1.1 200 OK\r\n" +
                    "Content-Type: text/plain\r\n" +
                    "Content-Length: " + fmt.Sprintf("%d", len(echoStr)) + "\r\n" +
                    "\r\n" +
                    echoStr
                _, err := conn.Write([]byte(response))
                if err != nil {
                    fmt.Println("Error writing response:", err)
                }
                return
            }
        }
    }

    // Default 404 Not Found response
    notFoundResponse := "HTTP/1.1 404 Not Found\r\n" +
        "Content-Length: 0\r\n" +
        "\r\n"

    _, err = conn.Write([]byte(notFoundResponse))
    if err != nil {
        fmt.Println("Error writing 404 response:", err)
    }
}
