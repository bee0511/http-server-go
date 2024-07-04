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
	Body      string
	UserAgent string
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	req, _ := parseStatus(scanner)
	fmt.Println(req)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(conn, "reading standard input:", err)
	}
	var response string
	switch path := req.Path; {
	case strings.HasPrefix(path, "/echo/"):
		content := strings.TrimPrefix(path, "/echo/")
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), content)
	case strings.HasPrefix(path, "/files/"):
		filePath := strings.TrimPrefix(path, "/files/")
		dir := os.Args[2]
		content, err := os.ReadFile(dir + "/" + filePath)
		if err != nil {
			response = getStatus(404, "Not Found") + "\r\n\r\n"
		} else {
			fmt.Println(string(content))
			response = fmt.Sprintf("%s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), string(content))
		}
	case path == "/user-agent":
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(req.UserAgent), req.UserAgent)
	case path == "/":
		response = getStatus(200, "OK") + "\r\n\r\n"
	default:
		response = getStatus(404, "Not Found") + "\r\n\r\n"
	}
	conn.Write([]byte(response))
}
func parseStatus(scanner *bufio.Scanner) (*HTTPRequest, error) {
	var req HTTPRequest = HTTPRequest{}
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
	}
	return &req, nil
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}
