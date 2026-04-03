package main

import (
	"fmt"

	sdk "github.com/MartinAbdrakhmanov/go-sdk"
)

func Handler(req sdk.Request) sdk.Response {
	return sdk.Response{
		Status: 200,
		Body:   fmt.Sprintf("test %s", req),
	}
}

func main() {
	sdk.Run(Handler)
}
