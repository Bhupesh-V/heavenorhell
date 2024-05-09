package instance

import (
	"log"
	"sync"

	"github.com/r3labs/sse/v2"
)

type instance struct {
	sseServer *sse.Server
}

var singleton = &instance{}
var once sync.Once

func Init() {
	once.Do(func() {
		server := sse.New()
		// disable replaying old events to new clients
		server.AutoReplay = false
		server.Headers = map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		}
		server.OnSubscribe = func(streamID string, sub *sse.Subscriber) {
			log.Println("A new client Subscribed", streamID)
		}
		server.OnUnsubscribe = func(streamID string, sub *sse.Subscriber) {
			log.Println("A client UnSubscribed", streamID)
		}
		singleton.sseServer = server
	})
}

// Validator returns the validator
func SSEServer() *sse.Server {
	return singleton.sseServer
}
