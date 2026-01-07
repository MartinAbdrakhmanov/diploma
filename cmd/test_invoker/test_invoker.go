package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/MartinAbdrakhmanov/diploma/internal/invoker"
)

func main() {
	out, logs, err := invoker.Invoke(context.Background(), ds.Function{ID: "test-123", Image: "docker.io/mini-faas/echo:323032362d30", Args: []string{"/handler"}}, []byte("hello from my own faas!"), time.Second*2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out), string(logs))
}
