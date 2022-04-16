package main

import (
	"io/ioutil"
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
	req, err := http.NewRequest(http.MethodGet, "http://worker-internal", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "Error creating request: " + err.Error()
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write([]byte(msg)); err != nil {
			log.Printf("Error writing response: %+v, %d", err, n)
		}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "Error sending request: " + err.Error()
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write([]byte(msg)); err != nil {
			log.Printf("Error writing response: %+v, %d", err, n)
		}
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "Error reading response: " + err.Error()
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write([]byte(msg)); err != nil {
			log.Printf("Error writing response: %+v, %d", err, n)
		}
		return
	}

	log.Printf("api received request: %s %s", r.Method, r.URL.Path)
	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("API server " + string(body)))
	if err != nil {
		log.Printf("api failed to write response: %+v, %d", err, n)
	}
}
