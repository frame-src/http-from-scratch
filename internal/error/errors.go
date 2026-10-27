package errors

import ("errors")


var (
	MalformedRequestLine = errors.New(
			"malformed HTTP request line")
	UnsupportedProtocol  = errors.New(
			"unsupported protocol (expected HTTP/1.1)")
	MissingHost          = errors.New(
			"missing Host header in HTTP/1.1 request")
	MalformedHeader      = errors.New(
			"HTTP malformed header")
)

var (
	ClientClosedConnection = errors.New(
			"client closed connection")
)


var (
	ErrWriteResponse = errors.New("Error WriteResponse")
)
