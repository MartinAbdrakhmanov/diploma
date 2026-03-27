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
	"github.com/MartinAbdrakhmanov/diploma/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	cont, err := container.New(closers)
	if err != nil {
		log.Fatalf("err while container init, %v", err)
	}
	defer func() {
		cont.Close()
		log.Println("All closers finished")
	}()

	errCh := make(chan error, 1)

	go func() {
		errCh <- runGW(ctx, cont)
	}()

	select {
	case err := <-errCh:
		log.Println("Gateway exited:", err)

	case <-ctx.Done():
		fmt.Println("Context cancelled")

		err := <-errCh
		log.Println("Gateway stopped:", err)
	}

	fmt.Println("Shutdown complete")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func runGW(ctx context.Context, cont *container.Container) error {
	gw, err := cont.GetApiGateway(ctx)
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir("./static"))))

	r.HandleFunc("/functions", gw.HandleRegister).Methods("POST")
	r.HandleFunc("/functions/{id}/invoke", gw.HandleInvoke).Methods("POST")
	r.HandleFunc("/functions/{id}/delete", gw.HandleDelete).Methods("POST")
	r.HandleFunc("/functions/{id}/stats", gw.HandleStats).Methods("GET")
	r.HandleFunc("/functions", gw.HandleList).Methods("GET")

	metrics.Init()
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsMiddleware(r),
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Println("Shutdown error:", err)
		}
	}()

	log.Printf("gateway listening on %s", srv.Addr)

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
