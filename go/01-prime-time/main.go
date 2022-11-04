package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

// To keep costs down, a hot new government department is contracting out its
// mission-critical primality testing to the lowest bidder. (That's you).

// Officials have devised a JSON-based request-response protocol. Each request
// is a single line containing a JSON object, terminated by a newline character
// ('\n', or ASCII 10). Each request begets a response, which is also a single
// line containing a JSON object, terminated by a newline character.

// After connecting, a client may send multiple requests in a single session.
// Each request should be handled in order.

// A conforming request object has the required field method, which must always
// contain the string "isPrime", and the required field number, which must
// contain a number. Any JSON number is a valid number, including
// floating-point values.

// Example request:

// {"method":"isPrime","number":123}
// A request is malformed if it is not a well-formed JSON object, if any
// required field is missing, if the method name is not "isPrime", or if the
// number value is not a number.

// Extraneous fields are to be ignored.

// A conforming response object has the required field method, which must
// always contain the string "isPrime", and the required field prime, which
// must contain a boolean value: true if the number in the request was prime,
// false if it was not.

// Example response:

// {"method":"isPrime","prime":false}
// A response is malformed if it is not a well-formed JSON object, if any
// required field is missing, if the method name is not "isPrime", or if the
// prime value is not a boolean.

// A response object is considered incorrect if it is well-formed but has an
// incorrect prime value. Note that non-integers can not be prime.

// Accept TCP connections.

// Whenever you receive a conforming request, send back a correct response, and
// wait for another request.

// Whenever you receive a malformed request, send back a single malformed
// response, and disconnect the client.

// Make sure you can handle at least 5 simultaneous clients.

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	for conn := range server.Connections() {
		server.Handle(conn, prime)
	}
	server.Wait()
}

func prime(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		req, err := NewRequest(scanner.Bytes())
		if err != nil {
			conn.Write(formatError(err))
			continue
		}
		resp, err := req.Process()
		if err != nil {
			conn.Write(formatError(err))
			continue
		}
		j, err := json.Marshal(resp)
		if err != nil {
			conn.Write(formatError(err))
			continue
		}
		conn.Write(append(j, '\n'))
	}
}

func formatError(err error) []byte {
	j, err := json.Marshal(map[string]string{"error": err.Error()})
	if err != nil {
		return []byte(`{"error":"internal server error"}\n`)
	}
	return append(j, '\n')
}
