package instance

import (
	"context"
	"heavenorhell/entities/notification"
	"heavenorhell/entities/store"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/go-github/github"
	"github.com/r3labs/sse/v2"
	"golang.org/x/oauth2"
)

type instance struct {
	sseServer      *sse.Server
	store          *store.Store
	Oauth2Client   *http.Client
	GitHubClient   *github.Client
	TelegramClient *notification.TelegramBot
}

var singleton = &instance{}
var once sync.Once

func Init() {
	once.Do(func() {
		server := sse.New()
		// disable replaying old events to new clients
		server.AutoReplay = false
		server.Headers = map[string]string{
			"Access-Control-Allow-Origin":  "heavenorhell.zyz",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		}
		server.OnSubscribe = func(streamID string, sub *sse.Subscriber) {
			// log.Println("A new client Subscribed", streamID)
		}
		server.OnUnsubscribe = func(streamID string, sub *sse.Subscriber) {
			// log.Println("A client UnSubscribed", streamID)
		}
		singleton.sseServer = server

		store_secret := os.Getenv("GITHUB_TOKEN")
		if store_secret == "" {
			panic("GITHUB_TOKEN is required")
		}
		store_id := os.Getenv("GIST_ID")
		if store_id == "" {
			panic("GIST_ID is required")
		}
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: store_secret},
		)
		tc := oauth2.NewClient(ctx, ts)

		client := github.NewClient(tc)
		// Get the existing gist
		log.Println("Getting initial store")
		gist, _, err := client.Gists.Get(ctx, store_id)
		if err != nil {
			panic(err)
		}

		// file := gist.Files[github.GistFilename(constants.STORE_FILE_NAME)]
		singleton.store = &store.Store{
			StoreID:    store_id,
			StoreToken: store_secret,
			DataFile:   gist,
		}
		singleton.Oauth2Client = tc
		singleton.GitHubClient = client

		telegramToken := os.Getenv("TELEGRAM_TOKEN")
		if telegramToken == "" {
			panic("TELEGRAM_TOKEN is required")
		}
		telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
		if telegramChatID == "" {
			panic("TELEGRAM_CHAT_ID is required")
		}
		singleton.TelegramClient = &notification.TelegramBot{
			Token:  telegramToken,
			ChatID: telegramChatID,
		}
	})
}

// Validator returns the validator
func SSEServer() *sse.Server {
	return singleton.sseServer
}

// Store returns the store
func Store() *store.Store {
	return singleton.store
}

// Oauth2Client returns the oauth2 client
func Oauth2Client() *http.Client {
	return singleton.Oauth2Client
}

// GitHubClient returns the github client
func GitHubClient() *github.Client {
	return singleton.GitHubClient
}

// TelegramClient returns the telegram client
func TelegramClient() *notification.TelegramBot {
	return singleton.TelegramClient
}
