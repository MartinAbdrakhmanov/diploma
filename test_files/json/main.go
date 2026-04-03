package main

import (
	sdk "github.com/MartinAbdrakhmanov/go-sdk"
)

func Handler(req sdk.Request) sdk.Response {
	data := req.Body
	if data == nil {
		data = make(map[string]interface{})
	}

	data["processed_by"] = "mini-faas"
	data["status"] = "success"

	return sdk.Response{
		Status: 200,
		Body:   data,
	}
}

func main() {
	sdk.Run(Handler)
}
