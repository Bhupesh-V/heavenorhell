package main

import (
	"encoding/json"
	"fmt"
	"heavenorhell/constants"
	"heavenorhell/entities/booking"
	"heavenorhell/instance"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/r3labs/sse/v2"
)

// Don't worry I will suffer in hell for this
var (
	heavenBookings = 0
	hellBookings   = 0
	mu             = &sync.Mutex{}
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type TemplateData struct {
	HeavenCount int
	HellCount   int
}

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
	mu.Lock()
	if afterlife == "Heaven" {
		// pick random message from HeavenMessages
		message = constants.HeavenMessages[rand.Intn(len(constants.HeavenMessages))]
		heavenBookings++
	} else {
		// pick random message from HellMessages
		message = constants.HellMessages[rand.Intn(len(constants.HellMessages))]
		hellBookings++
	}
	mu.Unlock()

	data := map[string]interface{}{
		"message": message,
		"counts": map[string]int{
			"heaven": heavenBookings,
			"hell":   hellBookings,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling data:", err)
		return
	}

	server := instance.SSEServer()
	server.Publish("messages", &sse.Event{
		Data: jsonData,
	})
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Write the response body
	response := fmt.Sprintf("You are booked for %s, your ticket id is <b>#%s</b>", afterlife, generateTicketID(afterlife))
	w.Write([]byte(response))
}

func main() {
	instance.Init()
	bookings, err := booking.GetBookings()
	if err != nil {
		log.Println("Error getting bookings")
		return
	}
	// server has been restarted, so we need to get the bookings from the gist
	if bookings.Heaven > 0 || bookings.Hell > 0 {
		heavenBookings = bookings.Heaven
		hellBookings = bookings.Hell
	}

	server := instance.SSEServer()
	server.CreateStream("messages")

	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.ServeHTTP)
	mux.HandleFunc("/choose-heaven", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Heaven")
	})
	mux.HandleFunc("/choose-hell", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Hell")
	})

	// mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			HeavenCount: heavenBookings,
			HellCount:   hellBookings,
		}

		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	addr := ":8080"
	log.Println("Starting server on", addr)

	// every 10 minutes, update the bookings
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			bookings := &booking.Bookings{
				Heaven: heavenBookings,
				Hell:   hellBookings,
			}
			err := bookings.Update()
			if err != nil {
				log.Println("Error updating bookings")
			}
		}
	}()

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatal(err)
	}
}
