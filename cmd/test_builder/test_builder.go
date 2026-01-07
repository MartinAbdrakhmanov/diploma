package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MartinAbdrakhmanov/diploma/internal/builder"
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
	image, err := builder.Build(ctx, "echo", files)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Built image:", image)
}
