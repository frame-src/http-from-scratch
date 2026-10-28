package http1

import (
	"bufio"
	// "bytes"
	"errors"
	"fmt"
	reqError "httpServer/internal/error"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
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
// According to the HTTP/1.1 specification (RFC 9112 §2.2),
// protocol elements like request lines and header fields must be delimited by CRLF (\r\n):
func readLineWithCRLF(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, bufio.ErrBufferFull) {
			return "", reqError.RequestHeaderFieldsTooLarge
		}
			return "", reqError.RequestEOF
		}
	if !strings.HasSuffix(line, "\r\n") {
				return "", reqError.RequestStatusLineNotTerminated
			}
	return string(line[:len(line)-2]), nil
}

func isValidMethod(m string) bool {
	var  methodsAllowed = [8]string {"GET","POST","PUT","DELETE","HEAD","OPTIONS","TRACE","CONNECT"}

	var match = false
	for  _, value := range methodsAllowed {
		if value == m {
			match = true
		}
	}
	return match
}

func parseRequestStatusLine(rl string ) (string, string, string, error) {
	parts := strings.Split(rl, " ")
	if len(parts) != 3 {
		return "", "", "", reqError.MalformedRequestLine
	}
	return parts[0], parts[1], parts[2], nil
}

func checkRequestStatusLine(method string, rawURI string, protocol string, req *http.Request) error {
	if !isValidMethod(method) {
		return reqError.NotValidMethod
	}
	parsedURL, err := url.ParseRequestURI(rawURI)
	if err != nil {
		return reqError.NotValidURI
	}
	major, minor, ok := http.ParseHTTPVersion(protocol)
	if !ok || major != 1 || minor != 1 {
		return reqError.UnsupportedProtocol
	}
	req.Method = method
	req.Proto = protocol
	req.ProtoMajor = major
	req.ProtoMinor = minor
	req.RequestURI = rawURI
	req.URL = parsedURL
	return nil
}

func parseRequestHeader(reader *bufio.Reader, header *http.Header) error {
	for {
		line, err := readLineWithCRLF(reader)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}
		if line == "" {
			break
		}
		colonIdx := strings.IndexByte(line, ':')
		if colonIdx == -1 {
			return reqError.MalformedHeader
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		header.Add(key, val)
	}
	return nil
}


func ParseRequest(reader *bufio.Reader) (*http.Request, error) {

	req := &http.Request{
		ContentLength: -1,
		Body:          http.NoBody,
	}
	
	requestLine, err := readLineWithCRLF(reader)
	if err != nil {
		return nil, fmt.Errorf("reading request line: %w", err)
	}
	
	method, rawURI, protocolVersion, err := parseRequestStatusLine(requestLine)
	if err != nil {
		return nil, reqError.MalformedRequestLine
	}
	err = checkRequestStatusLine(method,rawURI,protocolVersion, req)
	if err != nil {
		return nil, err
	}
	
	header := make(http.Header)
	err = parseRequestHeader(reader, &header)
	if err != nil {
		return nil, err
	}
	host := header.Get("Host")
	if host == "" {
		return nil, reqError.MissingHost
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

