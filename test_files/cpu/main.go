package main

import (
	sdk "github.com/MartinAbdrakhmanov/go-sdk"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func Handler(req sdk.Request) sdk.Response {
	n := 30

	if val, ok := req.Body["n"].(float64); ok {
		n = int(val)
	}

	result := fibonacci(n)

	return sdk.Response{
		Status: 200,
		Body: map[string]interface{}{
			"n":      n,
			"result": result,
		},
	}
}

func main() {
	sdk.Run(Handler)
}
