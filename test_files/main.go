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
