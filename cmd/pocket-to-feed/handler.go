package main

import (
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/feeds"
	pocket "github.com/motemen/go-pocket/api"
)

var _ http.Handler = &PocketFeedHandler{}

type PocketFeedHandler struct {
	cfg        Config
	cli        *pocket.Client
	defualtOpt *pocket.RetrieveOption
}

//nolint:funlen
func (h *PocketFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	retriveOpt := &pocket.RetrieveOption{}

	*retriveOpt = *h.defualtOpt
	queries := r.URL.Query()

	if tagParam, ok := queries["tag"]; ok {
		retriveOpt.Tag = tagParam[0]
	}

	result, err := h.cli.Retrieve(retriveOpt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	pocketItems := make([]pocket.Item, len(result.List))
	index := 0

	for _, pocketItem := range result.List {
		pocketItems[index] = pocketItem
		index++
	}

	sort.Slice(
		pocketItems,
		func(i, j int) bool {
			return time.Time(pocketItems[i].TimeAdded).After(time.Time(pocketItems[j].TimeAdded))
		})

	var latestAdded, latestUpdated time.Time

	items := make([]*feeds.Item, len(pocketItems))

	for i, result := range pocketItems {
		items[i] = &feeds.Item{
			Title:   result.Title(),
			Link:    &feeds.Link{Href: result.URL()},
			Created: time.Time(result.TimeAdded),
			Updated: time.Time(result.TimeUpdated),
		}

		if time.Time(result.TimeAdded).After(latestAdded) {
			latestAdded = time.Time(result.TimeAdded)
		}

		if time.Time(result.TimeUpdated).After(latestUpdated) {
			latestUpdated = time.Time(result.TimeUpdated)
		}
	}

	log.Printf("%+v, %+v\n", latestAdded, latestUpdated)

	feed := &feeds.Feed{
		Title:   h.cfg.Title,
		Link:    &feeds.Link{Href: h.cfg.URL},
		Created: latestAdded,
		Updated: latestUpdated,
		Items:   items,
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Println("toRSS:", err)

		return
	}

	if _, err := w.Write([]byte(rss)); err != nil {
		log.Println("write:", err)
	}
}

func createPocketFeedHandler(cfg Config, cli *pocket.Client, defaultOpt *pocket.RetrieveOption) *PocketFeedHandler {
	return &PocketFeedHandler{
		cfg:        cfg,
		cli:        cli,
		defualtOpt: defaultOpt,
	}
}
