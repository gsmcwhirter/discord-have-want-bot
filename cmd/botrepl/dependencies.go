package main

import (
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/go-kit/kit/log"

	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

type dependencies struct {
	logger  log.Logger
	db      *bolt.DB
	userAPI storage.UserAPI
}

func createDependencies(conf config) (d *dependencies, err error) {
	d = &dependencies{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	d.logger = logger

	d.db, err = bolt.Open(conf.Database, 0660, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return
	}

	d.userAPI, err = storage.NewBoltUserAPI(d.db)
	if err != nil {
		return
	}

	return
}

func (d *dependencies) Close() {
	if d.db != nil {
		d.db.Close() // nolint: errcheck
	}
}

func (d dependencies) Logger() log.Logger {
	return d.logger
}

func (d dependencies) UserAPI() storage.UserAPI {
	return d.userAPI
}
