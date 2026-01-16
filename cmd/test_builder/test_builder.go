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

`
	testGoMod = `
module testfunc

go 1.25.5

require github.com/MartinAbdrakhmanov/go-sdk v0.1.0

`

	testGoSum = `
github.com/MartinAbdrakhmanov/go-sdk v0.1.0 h1:+igibwaz6WpwL7UOWHby/5lhqHUW7+aobFZUYquiJ5s=
github.com/MartinAbdrakhmanov/go-sdk v0.1.0/go.mod h1:kyDT5UNltChVoNSjf476rbqhFx92hs3Az3CG4jWmX5k=

`
)

func main() {
	files := map[string][]byte{
		"main.go": []byte(testMainGo),
		"go.mod":  []byte(testGoMod),
		"go.sum":  []byte(testGoSum),
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
		Name:    "echo",
		Files:   files,
		Runtime: "wasm",
	}
	image, err := builder.Build(ctx, entry)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Built image:", image)
}
