package main

type primeError string

func (p primeError) Error() string {
	return string(p)
}

const (
	malformedRequestError = primeError("malformed request")
	missingMethodError    = primeError("missing method")
	invalidMethodError    = primeError("invalid method, expected 'isPrime'")
	missingNumberError    = primeError("missing number")
)
