package main

import (
	"log"
	"net/http"
	"os"

	"github.com/idkwhyureadthis/practice/internal/server"
)

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8000"
	}
	r := server.New()

	srv := http.Server{
		Handler: r,
		Addr:    ":" + port,
	}

	log.Println("server running at", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
