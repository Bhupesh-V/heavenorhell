package store

import (
	"github.com/google/go-github/github"
)

type Store struct {
	StoreID    string
	StoreToken string
	DataFile   *github.Gist
}
