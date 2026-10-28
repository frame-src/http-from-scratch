package http1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)


type Response struct {
	StatusCode    int
	Header        http.Header
	Body          io.Reader 	// payload (it is optional)
	ContentLength int64     	// -1 triggers chunked transfer
	Close         bool      	// "Connection status
}

type responseWriter struct {
	conn          net.Conn
	bw            *bufio.Writer
	req           *http.Request
	header        http.Header
	statusCode    int
	wroteHeader   bool
	contentLength int64
	bodyBuf       bytes.Buffer
	closeConn     bool
}


func WriteResponse(bw *bufio.Writer, req *http.Request, res *Response) error {
	if res.Header == nil {
		res.Header = make(http.Header)
	}

	if res.Header.Get("Date") == "" {
		res.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	clientWantsClose := strings.EqualFold(req.Header.Get("Connection"), "close")
	shouldClose := res.Close || clientWantsClose
	if shouldClose {
		res.Header.Set("Connection", "close")
	} else if req.ProtoMinor == 0 { // Because of the HTTP/1.0 requires explicit keep-alive
		res.Header.Set("Connection", "keep-alive")
	}

	// 1xx, 204, 304, and HEAD responses must not include a body
	isNoBodyStatus := (res.StatusCode >= 100 && res.StatusCode < 200) ||
		res.StatusCode == http.StatusNoContent ||
		res.StatusCode == http.StatusNotModified
	isHeadReq := req.Method == http.MethodHead

	// Write Status-Line: HTTP/1.1 <StatusCode> <StatusText>\r\n
	statusText := http.StatusText(res.StatusCode)
	if statusText == "" {
		statusText = "status"
	}
	if _, err := fmt.Fprintf(bw, "HTTP/1.1 %d %s\r\n", res.StatusCode, statusText); err != nil {
		return err
	}

	// Content-Length or Chunked with -1)
	var bodyWriter io.Writer = bw
	var chunkedWriter io.WriteCloser

	if !isNoBodyStatus {
		if res.ContentLength >= 0 {
			res.Header.Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))
		} else if res.Body != nil {
			res.Header.Set("Transfer-Encoding", "chunked")
			chunkedWriter = httputil.NewChunkedWriter(bw)
			bodyWriter = chunkedWriter
		} else {
			res.Header.Set("Content-Length", "0")
		}
	}

	err := res.Header.Write(bw); 
	if err != nil {
		return fmt.Errorf("writing headers: %w", err)
	}
	if !isNoBodyStatus && !isHeadReq && res.Body != nil {
		if _, err := io.Copy(bodyWriter, res.Body); err != nil {
			return fmt.Errorf("writing body: %w", err)
		}
		// need to add the terminal 0\r\n\r\n
		if chunkedWriter != nil {
			if err := chunkedWriter.Close(); err != nil {
				return fmt.Errorf("closing chunked writer: %w", err)
			}
		}
	}
	return bw.Flush()
}


func NewResponse(conn net.Conn) *Response {
	panic("NewResponse: Not Implemented")
}

func (r *Response) HeaderToHttp() http.Header {
	return r.Header
}

func (r *Response) Write(data []byte) (n int, err error) {
	panic("Response.Write: Not Implemented")
}

func (r *Response) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}

