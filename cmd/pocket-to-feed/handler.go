package main

import (
	"log"
	"net/http"
	"net/url"
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

//nolint:funlen,cyclop
func (h *PocketFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !(r.Method == http.MethodGet || r.Method == http.MethodHead) {
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	feed := &feeds.Feed{
		Title: h.cfg.Title,
		Link:  &feeds.Link{Href: h.cfg.URL},
	}

	result, err := h.retrievePocket(r.URL.Query())
	if err != nil || result == nil {
		if err := feed.WriteRss(w); err != nil {
			log.Println("toRSS:", err)

			return
		}

		return
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

	feed.Created = latestAdded
	feed.Updated = latestUpdated
	feed.Items = items

	if err := feed.WriteRss(w); err != nil {
		log.Println("toRSS:", err)

		return
	}
}

func (h *PocketFeedHandler) retrievePocket(queries url.Values) (*pocket.RetrieveResult, error) {
	retriveOpt := &pocket.RetrieveOption{}
	*retriveOpt = *h.defualtOpt

	if tagParam, ok := queries["tag"]; ok {
		retriveOpt.Tag = tagParam[0]
	}

	//nolint:wrapcheck
	return h.cli.Retrieve(retriveOpt)
}

func createPocketFeedHandler(cfg Config, cli *pocket.Client, defaultOpt *pocket.RetrieveOption) *PocketFeedHandler {
	return &PocketFeedHandler{
		cfg:        cfg,
		cli:        cli,
		defualtOpt: defaultOpt,
	}
}
