package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("Signal received, cancelling context...")
		cancel()
	}()

	closers := make([]func(), 0, 2)
	cont := container.New(closers)
	defer cont.Close()

	errCh := make(chan error, 1)

	go func() {
		errCh <- runGW(ctx, cont)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	case err := <-errCh:
		log.Println("Gateway exited:", err)
	}

	fmt.Println("Shutdown complete")
}

func runGW(ctx context.Context, cont *container.Container) error {
	gw, err := cont.GetApiGateway(ctx)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/functions", gw.HandleRegister).Methods("POST")
	r.HandleFunc("/functions/{id}/invoke", gw.HandleInvoke).Methods("POST")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("gateway listening on %s", srv.Addr)
	return srv.ListenAndServe()
}
