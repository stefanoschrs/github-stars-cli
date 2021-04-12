package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/stefanoschrs/github-stars-cli/types"

	"github.com/dgraph-io/badger/v2"
)

type DB struct {
	*badger.DB
}

func Init() (db DB, err error) {
	opt := badger.DefaultOptions("").WithInMemory(true)
	if os.Getenv("STORAGE_FILE") != "" {
		opt = badger.DefaultOptions(os.Getenv("STORAGE_FILE"))
	}
	opt.Logger = nil

	db.DB, err = badger.Open(opt)
	if err != nil {
		return
	}

	return
}

func (db DB) GetUserRepos(username string) (repos *[]types.Repo, err error) {
	err = db.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte(username))
		if err != nil {
			return
		}

		return item.Value(func(val []byte) (err error) {
			repos = &[]types.Repo{}
			return json.Unmarshal(val, repos)
		})
	})
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return
		}

		err = nil
	}

	return
}

func (db DB) SaveUserRepos(username string, repos []types.Repo) (err error) {
	return db.Update(func(txn *badger.Txn) (err error) {
		reposStr, err := json.Marshal(repos)
		if err != nil {
			return
		}

		return txn.Set([]byte(username), reposStr)
	})
}
