package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	pocket "github.com/motemen/go-pocket/api"
)

const (
	HTTPReadTimeout  = 30 * time.Second
	HTTPWriteTimeout = 30 * time.Second
)

func main() {
	var (
		pcfg PocketConfig
		cfg  Config
	)

	err := envconfig.Process("pocket", &pcfg)
	if err != nil {
		log.Println("parse PocketConfig: ", err)
		os.Exit(1)
	}

	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Println("parse Config: ", err)
		os.Exit(1)
	}

	cli := pocket.NewClient(pcfg.ConsumerKey, pcfg.AccessToken)
	defaultOpt := &pocket.RetrieveOption{
		State:      pocket.StateUnread,
		Sort:       pocket.SortNewest,
		DetailType: pocket.DetailTypeSimple,
	}
	handler := createPocketFeedHandler(cfg, cli, defaultOpt)

	mux := http.NewServeMux()
	mux.Handle("/feed", handler)

	svr := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: HTTPWriteTimeout,
	}

	log.Println("starting server on :8080")
	log.Fatal(svr.ListenAndServe())
}
