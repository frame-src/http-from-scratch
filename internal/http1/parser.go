package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	reqError "httpServer/internal/error"
)

/*notes
HTTP does not place a predefined limit on the length of a request-line
a method longer than any that it implements 
->	SHOULD respond with a 501 (Not Implemented) status code.
A server that receives a request-target longer than any URI it wishes to parse
->	MUST respond with a 414 (URI Too Long)

t is RECOMMENDED that all HTTP senders and recipients support, at a minimum, request-line lengths of 8000 octets

A request is structured in 3 parts:

the mandatory first line of the request
the optional headers
the request body, also optional
---------------------------
REQUEST-LINE -> mandatory.
[HEADERS] -> optional

[BODY] -> optional
---------------------------
*/

// readCRLFLine reads bytes until \n and strips trailing \r\n or \n.
func readLineWithCRLF(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	return line, nil
}

func isValidMethod(m string) bool {
	var  METHOD_ALLOWED = [8]string {"GET","POST","PUT","DELETE","HEAD","OPTIONS","TRACE","CONNECT"}

	var match = false
	for _,value range METHOD_ALLOWED {
		if value == m {
			match = true
		}
	}
	return false
}

func parseRequestStatusLine(rl *bufio.Reader) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, reqError.MalformedRequestLine
	}
	parts[0], parts[1], parts[2]
}

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {

	requestLine, err := readLineWithCRLF(reader)
	if err != nil {
		return nil, fmt.Errorf("reading request line: %w", err)
	}
	
	method, rawURI, protocolVersion := parseRequestStatusLine()
	if !isValidMethod(method) {
		return er
	}
	major, minor, ok := http.ParseHTTPVersion(protocolVersion)
	if !ok || major != 1 || minor != 1 {
		return nil, reqError.UnsupportedProtocol
	}

	//HEADER
	header := make(http.Header)
	for {
		line, err := readLineWithCRLF(reader)
		if err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}
		if line == "" {
			break
		}

		colonIdx := strings.IndexByte(line, ':')
		if colonIdx == -1 {
			return nil, reqError.MalformedHeader
		}

		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		header.Add(key, val)
	}

	host := header.Get("Host")
	if host == "" {
		return nil, reqError.MissingHost
	}

	parsedURL, err := url.ParseRequestURI(rawURI)
	if err != nil {
		return nil, fmt.Errorf("parsing URI: %w", err)
	}

	req := &http.Request{
		Method:        method,
		URL:           parsedURL,
		Proto:         protocolVersion,
		ProtoMajor:    major,
		ProtoMinor:    minor,
		Header:        header,
		Host:          host,
		RequestURI:    rawURI,
		ContentLength: -1,
		Body:          http.NoBody,
	}

	transferModality := strings.ToLower(header.Get("Transfer-Encoding"))
	if strings.Contains(transferModality, "chunked") {
		req.Body = io.NopCloser(httputil.NewChunkedReader(reader))
	} else if clStr := header.Get("Content-Length"); clStr != "" {
		cl, err := strconv.ParseInt(clStr, 10, 64)
		if err != nil || cl < 0 {
			return nil, errors.New("invalid Content-Length")
		}
		req.ContentLength = cl
		if cl > 0 {
			req.Body = io.NopCloser(io.LimitReader(reader, cl))
		}
	}

	return req, nil
}

