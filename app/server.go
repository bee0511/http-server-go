package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type HTTPRequest struct {
	Method    string
	Path      string
	Headers   map[string]string
	Body      string
	UserAgent string
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	scanner := bufio.NewScanner(conn)
	req, _ := parseStatus(scanner)
	// fmt.Println(req.Headers)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(conn, "reading standard input:", err)
	}
	var response string
	switch path := req.Path; {
	case strings.HasPrefix(path, "/echo/"):
		content := strings.TrimPrefix(path, "/echo/")
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), content)
	case path == "/user-agent":
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(req.UserAgent), req.UserAgent)
	case path == "/":
		response = getStatus(200, "OK") + "\r\n\r\n"
	default:
		response = getStatus(404, "Not Found") + "\r\n\r\n"
	}
	conn.Write([]byte(response))
	// fmt.Fprintln(conn)
	conn.Close()
}
func parseStatus(scanner *bufio.Scanner) (*HTTPRequest, error) {
	var req HTTPRequest = HTTPRequest{}
	req.Headers = make(map[string]string)
	for i := 0; scanner.Scan(); i++ {
		if i == 0 {
			parts := strings.Split(scanner.Text(), " ")
			req.Method = parts[0]
			req.Path = parts[1]
			continue
		}
		headers := strings.Split(scanner.Text(), ": ")
		if len(headers) < 2 {
			req.Body = headers[0]
			break
		}
		if headers[0] == "User-Agent" {
			req.UserAgent = headers[1]
		}
		req.Headers[headers[0]] = headers[1]
	}
	return &req, nil
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}
