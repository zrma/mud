package main

import (
	"log"
	"net/http"
	"os"
)

const defaultAddr = ":8080"

func main() {
	addr := defaultAddr
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Server listening on port %s", addr)

	http.HandleFunc("/", home)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server listening error: %+v", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Printf("worker received request: %s %s", r.Method, r.URL.Path)
	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("this is worker server"))
	if err != nil {
		log.Printf("worker failed to write response: %+v, %d", err, n)
	}
}
