package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/google/uuid"
)

const (
	testMainGo = `
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	data, _ := io.ReadAll(os.Stdin)
	fmt.Printf("echo: %s", string(data))
}
`
	testGoMod = `
module testfunc

go 1.22`
)

func main() {
	files := map[string][]byte{
		"main.go": []byte(testMainGo),
		"go.mod":  []byte(testGoMod),
	}
	ctx := context.Background()
	closers := make([]func(), 0, 2)
	cont := container.New(closers)
	defer cont.Close()
	registry, err := cont.GetFunctionRegistry(ctx)
	if err != nil {
		log.Fatal(err)
	}
	userID := uuid.New()

	id, err := registry.Register(ctx, ds.Entry{UserId: userID.String(), Name: "test-registry-123", Files: files, Runtime: ds.DockerRuntime})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)

	fn, err := registry.Get(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fn)
}
