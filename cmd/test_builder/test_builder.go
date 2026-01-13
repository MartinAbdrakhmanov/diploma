package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
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
	builder, err := cont.GetBuilderSvc(ctx)
	if err != nil {
		log.Fatal(err)
	}
	entry := ds.Entry{
		Name:  "echo",
		Files: files,
	}
	image, err := builder.Build(ctx, entry)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Built image:", image)
}
