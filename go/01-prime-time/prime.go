package main

import (
	"encoding/json"
	"math/big"
)

type Request struct {
	Method *string  `json:"method"`
	Number *float64 `json:"number"`
}

type Response struct {
	Method string  `json:"method"`
	Number float64 `json:"number"`
	Prime  bool    `json:"isPrime"`
}

func NewRequest(input []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, malformedRequestError
	}
	return &req, nil
}

func (r *Request) valid() error {
	if r.Method == nil {
		return missingMethodError
	}
	if *r.Method != "isPrime" {
		return invalidMethodError
	}
	if r.Number == nil {
		return missingNumberError
	}
	return nil
}

func (r *Request) Process() (*Response, error) {
	if err := r.valid(); err != nil {
		return nil, err
	}
	var num int64
	if float64(int64(*r.Number)) == *r.Number {
		num = int64(*r.Number)
	}
	return &Response{
		Method: *r.Method,
		Number: *r.Number,
		Prime:  big.NewInt(num).ProbablyPrime(0),
	}, nil
}
