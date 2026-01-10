package main

import (
	"context"
	"log"
	"net/http"

	"github.com/MartinAbdrakhmanov/diploma/internal/container"
	"github.com/gorilla/mux"
)

func main() {

	ctx := context.Background()
	closers := make([]func(), 0, 2)
	cont := container.New(closers)
	defer cont.Close()
	gw, err := cont.GetApiGateway(ctx)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/functions", gw.HandleRegister).Methods("POST")
	r.HandleFunc("/functions/{id}/invoke", gw.HandleInvoke).Methods("POST")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	log.Printf("gateway listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
