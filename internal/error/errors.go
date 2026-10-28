package errors

import ("errors")


var (
	MalformedRequestLine = errors.New(
			"malformed HTTP request line")
	NotValidMethod = errors.New(
			"not a valid method")
	NotValidURI = errors.New(
			"not a valid URI")
	UnsupportedProtocol  = errors.New(
			"unsupported protocol (expected HTTP/1.1)")
	MissingHost          = errors.New(
			"missing Host header in HTTP/1.1 request")
	MalformedHeader      = errors.New(
			"HTTP malformed header")
	RequestHeaderFieldsTooLarge = errors.New(
			"header line too long")
	RequestStatusLineNotTerminated = errors.New (
			"line not terminated with CRLF")
	RequestEOF = errors.New (
		"io.EOF or unexpected connection drops")
)

var (
	ClientClosedConnection = errors.New(
			"client closed connection")
)


var (
	ErrWriteResponse = errors.New("Error WriteResponse")
)
