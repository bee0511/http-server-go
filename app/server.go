package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
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
	req, _ := parseStatus(conn)
	fmt.Println(req)

	var response string
	switch path := req.Path; {
	case strings.HasPrefix(path, "/echo/"):
		content := strings.TrimPrefix(path, "/echo/")
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), content)
	case strings.HasPrefix(path, "/files/"):
		filePath := strings.TrimPrefix(path, "/files/")
		dir := os.Args[2]
		if req.Method == "POST" {
            err := os.WriteFile(dir + "/" + filePath, []byte(req.Body), 0644)
            if err != nil{
                panic(err)
            }
			response = getStatus(201, "Created") + "\r\n\r\n"
			break
		}
		content, err := os.ReadFile(dir + "/" + filePath)
		if err != nil {
			response = getStatus(404, "Not Found") + "\r\n\r\n"
		} else if req.Method == "GET" {
			fmt.Println("Content of the file: ", string(content))
			response = fmt.Sprintf("%s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), string(content))
		}
	case path == "/user-agent":
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(req.Headers["User-Agent"]), req.Headers["User-Agent"])
	case path == "/":
		response = getStatus(200, "OK") + "\r\n\r\n"
	default:
		response = getStatus(404, "Not Found") + "\r\n\r\n"
	}
	conn.Write([]byte(response))
}
func parseStatus(conn net.Conn) (*HTTPRequest, error) {
	var req HTTPRequest = HTTPRequest{}
	req.Headers = make(map[string]string)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	request := string(buf[:n])
	lines := strings.Split(request, "\r\n")

	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		return nil, fmt.Errorf("invalid request line")
	}
	req.Method = requestLine[0]
	req.Path = requestLine[1]

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			req.Headers[headerParts[0]] = headerParts[1]
		}
	}

	if lenStr, ok := req.Headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length")
		}
		bodyIndex := bytes.Index(buf, []byte("\r\n\r\n")) + 4
		req.Body = string(buf[bodyIndex : bodyIndex+contentLength])
	}

	return &req, nil
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}
