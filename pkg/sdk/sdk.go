package sdk

import (
	"encoding/json"
	"io"
	"os"
)

type Request struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

func Run(handler func(Request) Response) {
	input, _ := io.ReadAll(os.Stdin)

	var req Request
	json.Unmarshal(input, &req)

	resp := handler(req)

	out, _ := json.Marshal(resp)
	os.Stdout.Write(out)
}
