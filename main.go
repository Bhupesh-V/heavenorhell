package main

import (
	"fmt"
	"heavenorhell/constants"
	"heavenorhell/instance"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateTicketID(afterlife string) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 13)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", afterlife, string(b))
}

func logHTTPRequest(w http.ResponseWriter, r *http.Request, afterlife string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "isTicketBooked",
		Value: "true",
	})

	var message string
	if afterlife == "Heaven" {
		// pick random message from HeavenMessages
		message = constants.HeavenMessages[rand.Intn(len(constants.HeavenMessages))]
	} else {
		// pick random message from HellMessages
		message = constants.HellMessages[rand.Intn(len(constants.HellMessages))]
	}

	server := instance.SSEServer()
	server.Publish("messages", &sse.Event{
		Data: []byte(message),
	})
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Write the response body
	response := fmt.Sprintf("You are booked for %s, your ticket id is <b>#%s</b>", afterlife, generateTicketID(afterlife))
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
