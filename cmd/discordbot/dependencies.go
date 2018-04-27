package main

import (
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/go-kit/kit/log"

	"github.com/gsmcwhirter/eso-discord/pkg/httpclient"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type dependencies struct {
	logger     log.Logger
	db         *bolt.DB
	userAPI    storage.UserAPI
	httpClient httpclient.HTTPClient
	wsClient   wsclient.WSClient
}

func createDependencies(conf config) (d *dependencies, err error) {
	d = &dependencies{}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
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

	d.httpClient = httpclient.NewHTTPClient(d)
	d.wsClient = wsclient.NewWSClient(d, wsclient.Options{})

	return
}

func (d *dependencies) Close() {
	if d.db != nil {
		d.db.Close()
	}

	if d.wsClient != nil {
		d.wsClient.Close()
	}
}

func (d dependencies) Logger() log.Logger {
	return d.logger
}

func (d dependencies) HTTPClient() httpclient.HTTPClient {
	return d.httpClient
}

func (d dependencies) WSClient() wsclient.WSClient {
	return d.wsClient
}

func (d dependencies) UserAPI() storage.UserAPI {
	return d.userAPI
}
