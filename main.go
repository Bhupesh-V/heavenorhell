package main

import (
	"heavenorhell/instance"
	"log"
	"net/http"

	"github.com/r3labs/sse/v2"
)

func logHTTPRequest(w http.ResponseWriter, r *http.Request, afterlife string) {
	log.Printf("Got trigger request. Sending SSE")

	server := instance.SSEServer()
	server.Publish("messages", &sse.Event{
		Data: []byte("Someone choosen " + afterlife),
	})
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Write the response body
	response := "You have chosen " + afterlife
	w.Write([]byte(response))
}

func main() {
	instance.Init()
	server := instance.SSEServer()

	server.CreateStream("messages")

	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.ServeHTTP)
	// mux.HandleFunc("/trigger", logHTTPRequest)
	mux.HandleFunc("/choose-heaven", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Heaven")
	})
	mux.HandleFunc("/choose-hell", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Hell")
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))
	addr := ":8080"
	log.Println("Starting server on", addr)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatal(err)
	}

}