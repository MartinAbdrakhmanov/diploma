package main

import (
	"context"
	"log"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	closers := make([]func(), 0)
	cont, err := container.New(closers)
	if err != nil {
		log.Fatalf("err while container init, %v", err)
	}
	defer func() {
		cont.Close()
		log.Println("All closers finished")
	}()

	cleaner, err := cont.Cleaner(ctx)
	if err != nil {
		log.Fatalf("err while cleaner init, %v", err)
	}

	cleaner.Run(ctx)
}
