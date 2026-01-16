package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
)

func main() {
	closers := make([]func(), 0, 2)
	cont := container.New(closers)
	defer cont.Close()
	invoker, err := cont.GetInvokerSvc(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	input := []byte(`
{
  "method": "POST",
  "path": "/",
  "headers": {
    "content-type": "application/json"
  },
  "body": {
    "message": "hello from wasm"
  }
}
`)
	// out, logs, err := invoker.Invoke(context.Background(), ds.Function{ID: "test-123", Image: "docker.io/mini-faas/echo:323032362d30", Args: []string{"/handler"}}, []byte("hello from my own faas!"), time.Second*2)
	out, logs, err := invoker.Invoke(context.Background(), ds.Function{ID: "test-123", Runtime: "wasm", WasmPath: "/var/lib/mini-faas/wasm/echo.wasm", Args: []string{"/handler"}}, input, time.Second*2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out), string(logs))
}
