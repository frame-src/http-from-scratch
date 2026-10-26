package http1

import (
	"bufio"
	"net/http"
)

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	panic("ParseRequest: Not Implemented")
	data, err = reader.Read(data)
    if err != nil {
        fmt.Println(err)
        return
    }
	return data
}

/*notes
HTTP does not place a predefined limit on the length of a request-line
a method longer than any that it implements 
->	SHOULD respond with a 501 (Not Implemented) status code.
A server that receives a request-target longer than any URI it wishes to parse
->	MUST respond with a 414 (URI Too Long)

t is RECOMMENDED that all HTTP senders and recipients support, at a minimum, request-line lengths of 8000 octets
