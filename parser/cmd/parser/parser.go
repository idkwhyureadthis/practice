package main

import (
	"log"
	"net/http"

	"github.com/idkwhyureadthis/practice/internal/server"
)

func main() {
	r := server.New()

	srv := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("server running at", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
