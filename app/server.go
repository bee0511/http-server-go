package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
    "compress/gzip"
	"strings"
)

type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type HTTPResponse struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Content    string
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

	response := generateResponse(req)
	conn.Write([]byte(getResponse(response)))
}
func generateResponse(req *HTTPRequest) *HTTPResponse {
	res := &HTTPResponse{
		Headers: make(map[string]string),
	}
	switch {
	case strings.HasPrefix(req.Path, "/echo/"):
		content := strings.TrimPrefix(req.Path, "/echo/")
		res.StatusCode = 200
		res.StatusText = "OK"
		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = strconv.Itoa(len(content))
		encoding, ok := req.Headers["Accept-Encoding"]
		if ok && strings.Contains(encoding, "gzip") {
			res.Headers["Content-Encoding"] = "gzip"
            var buffer bytes.Buffer
            w := gzip.NewWriter(&buffer)
            w.Write([]byte(content))
            w.Close()
            res.Content = buffer.String()
            break
		}
		res.Content = content
	case strings.HasPrefix(req.Path, "/files/"):
		filePath := strings.TrimPrefix(req.Path, "/files/")
		dir := os.Args[2]
		if req.Method == "POST" {
			err := os.WriteFile(dir+"/"+filePath, []byte(req.Body), 0644)
			if err != nil {
				panic(err)
			}
			res.StatusCode = 201
			res.StatusText = "Created"
			res.Headers["Content-Type"] = "text/plain"
			break
		}
		content, err := os.ReadFile(dir + "/" + filePath)
		if err != nil {
			res.StatusCode = 404
			res.StatusText = "Not Found"
		} else if req.Method == "GET" {
			res.StatusCode = 200
			res.StatusText = "OK"
			res.Headers["Content-Type"] = "application/octet-stream"
			res.Headers["Content-Length"] = strconv.Itoa(len(content))
			res.Content = string(content)
		}
	case req.Path == "/user-agent":
		userAgent := req.Headers["User-Agent"]
		res.StatusCode = 200
		res.StatusText = "OK"
		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
		res.Content = userAgent
	case req.Path == "/":
		res.StatusCode = 200
		res.StatusText = "OK"
		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = "0"
	default:
		res.StatusCode = 404
		res.StatusText = "Not Found"
	}
	return res
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

func getResponse(res *HTTPResponse) string {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.StatusCode, res.StatusText)
	headers := ""

	// Add headers from the map
	for key, value := range res.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	headers += "\r\n" // End of headers
	return statusLine + headers + res.Content
}
