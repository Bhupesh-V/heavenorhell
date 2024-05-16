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
	"net/url"
	"os"
	"strings"
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
	HeavenCount  int
	HellCount    int
	FaviconEmoji string
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
	name := strings.TrimSpace(r.FormValue("name"))

	var message string
	mu.Lock()
	if afterlife == "Heaven" {
		// pick random message from HeavenMessages
		message = fmt.Sprintf(constants.HeavenMessages[rand.Intn(len(constants.HeavenMessages))], name)
		heavenBookings++
	} else {
		// pick random message from HellMessages
		message = fmt.Sprintf(constants.HellMessages[rand.Intn(len(constants.HellMessages))], name)
		hellBookings++
	}
	mu.Unlock()

	go func() {
		sendTelegramUpdate(fmt.Sprintf("%s has been booked for %s", name, afterlife))
	}()

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
	ticketID := generateTicketID(afterlife)
	response := fmt.Sprintf("%s you are booked for %s, your ticket id is <b class=\"ticketText %s\">#%s</b>", name, afterlife, afterlife, ticketID)
	http.SetCookie(w, &http.Cookie{
		Name:  "isTicketBooked",
		Value: "true",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "ticketID",
		Value: ticketID,
	})

	w.Write([]byte(response))
}

func sendTelegramUpdate(message string) {
	telegramInstance := instance.TelegramClient()
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramInstance.Token)
	data := url.Values{
		"chat_id": {telegramInstance.ChatID},
		"text":    {message},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		log.Println("Error sending Telegram message:", err)
	}
	defer resp.Body.Close()
}

func allowDomainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Replace 'alloweddomain.com' with your domain
		log.Println(r.Host)
		if r.Host == "heavenorhell.xyz" || r.Host == os.Getenv("IP_ADDRESS") {
			next.ServeHTTP(w, r)
		} else {
			fmt.Println("Forbidden request from IP:", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func main() {
	instance.Init()
	bookings, err := booking.GetBookings()
	if err != nil {
		log.Println("Error getting bookings")
		return
	}
	log.Printf("Current bookings: %#v\n", bookings)
	// server has been restarted, so we need to get the bookings from the gist
	heavenBookings = bookings.Heaven
	hellBookings = bookings.Hell

	server := instance.SSEServer()
	server.CreateStream("messages")

	mux := http.NewServeMux()
	wrappedMux := allowDomainMiddleware(mux)
	mux.HandleFunc("/events", server.ServeHTTP)
	mux.HandleFunc("/choose-heaven", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Heaven")
	})
	mux.HandleFunc("/choose-hell", func(w http.ResponseWriter, r *http.Request) {
		logHTTPRequest(w, r, "Hell")
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})

	// mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		emojis := []string{"👹", "😇"}
		rand.Seed(time.Now().UnixNano())
		data := TemplateData{
			HeavenCount:  heavenBookings,
			HellCount:    hellBookings,
			FaviconEmoji: emojis[rand.Intn(len(emojis))],
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

	// every 5 minutes, try to update the bookings
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
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

	err = http.ListenAndServe(addr, wrappedMux)
	if err != nil {
		log.Fatal(err)
	}
}
