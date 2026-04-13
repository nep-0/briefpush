package main

import (
	"briefpush/ai"
	"briefpush/config"
	"briefpush/feed"
	"briefpush/notify"
	"briefpush/store"
)

func main() {
	config, err := config.LoadConfig("config.json")
	if err != nil {
		panic(err)
	}

	store := store.NewJsonStore(config.JsonFilePath)
	AI := ai.NewLLM(config.BaseURL, config.ApiKey, config.Model)
	notifier, err := notify.NewDispatcher(config.Notification)
	if err != nil {
		panic(err)
	}

	feedManager := feed.NewFeedManager(store, AI, notifier, config.ReportHour, config.Feeds)

	feedManager.Start()
	select {}
}
