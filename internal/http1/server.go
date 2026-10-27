package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	errorHandler "httpServer/internal/error"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen: %w", err)
		}
		defer listener.Close()
	
		if handler == nil {
			handler = http.DefaultServeMux
		}
	
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return fmt.Errorf("accept: %w", err)
			}
			go HandleConnection(conn, handler)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(15 * time.Second))

		req, err := ParseRequest(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errorHandler.ClientClosedConnection
			}
			badRequest := &Response{
				StatusCode:    http.StatusBadRequest,
				ContentLength: 0,
				Close:         true,
			}
			_ = WriteResponse(writer, &http.Request{Method: "GET"}, badRequest)
			return nil
		}

		res := &Response{
			StatusCode:    http.StatusOK,
			Header:        make(http.Header),
			Body:          strings.NewReader("Hello, world!"),
			ContentLength: int64(len("Hello, world!")),
			Close:         false,
		}
		res.Header.Set("Content-Type", "text/plain; charset=utf-8")

		if err := WriteResponse(writer, req, res); err != nil {
			return errorHandler.ErrWriteResponse
		}

		// If either side signaled close, terminate the TCP connection
		if res.Close || strings.EqualFold(req.Header.Get("Connection"), "close") {
			return nil
		}
	}
}