package booking

import (
	"context"
	"encoding/json"
	"heavenorhell/constants"
	"heavenorhell/instance"
	"log"

	"github.com/google/go-github/github"
)

type Bookings struct {
	Hell   int `json:"hell"`
	Heaven int `json:"heaven"`
}

func (b *Bookings) Update() error {
	ctx := context.Background()
	storeInstance := instance.Store()
	client := instance.GitHubClient()

	// Marshal the struct back into JSON
	bookingsJson, err := json.Marshal(b)
	if err != nil {
		log.Println("Error marshalling emails")
		return err
	}

	gist := storeInstance.DataFile

	file := gist.Files[github.GistFilename(constants.STORE_FILE_NAME)]
	file.Content = github.String(string(bookingsJson))
	gist.Files[github.GistFilename(constants.STORE_FILE_NAME)] = file

	// Update the gist
	_, _, err = client.Gists.Edit(ctx, storeInstance.StoreID, gist)
	if err != nil {
		log.Println("Error updating hellorheaven bookings gist")
		return err
	}
	return nil
}

func GetBookings() (*Bookings, error) {
	storeInstance := instance.Store()
	gist := storeInstance.DataFile

	// unmarshal the existing JSON content into a Go struct
	var bookings Bookings
	file := gist.Files[github.GistFilename(constants.STORE_FILE_NAME)]
	err := json.Unmarshal([]byte(*file.Content), &bookings)
	if err != nil {
		log.Println("Error unmarshalling gist content")
		return nil, err
	}

	return &bookings, nil
}
