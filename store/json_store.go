package store

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/mmcdole/gofeed"
)

type JsonStore struct {
	JsonFile string
}

type JsonFeed struct {
	Key  string
	Name string
	URL  string
	Feed *gofeed.Feed
}

func NewJsonStore(jsonFile string) *JsonStore {
	return &JsonStore{
		JsonFile: jsonFile,
	}
}

func (s *JsonStore) AddFeed(key string, name string, url string, feed *gofeed.Feed) error {
	feeds, err := s.loadFeeds()
	if err != nil {
		return err
	}
	feeds[key] = &JsonFeed{
		Key:  key,
		Name: name,
		URL:  url,
		Feed: feed,
	}
	return s.saveFeeds(feeds)
}

func (s *JsonStore) UpdateFeed(key string, feed *gofeed.Feed) error {
	feeds, err := s.loadFeeds()
	if err != nil {
		return err
	}
	if feeds[key].Feed == nil {
		feeds[key].Feed = feed
	} else {
		previousItems := feeds[key].Feed.Items
		newItems := []*gofeed.Item{}
		if previousItems != nil {
			for _, item := range feed.Items {
				isNew := true
				for _, prevItem := range previousItems {
					if item.GUID == prevItem.GUID {
						isNew = false
					}
				}
				if isNew {
					newItems = append(newItems, item)
				}
			}
		}
		feeds[key].Feed.Items = append(newItems, feeds[key].Feed.Items...)
	}
	return s.saveFeeds(feeds)
}

func (s *JsonStore) GetFeed(key string) (feed *gofeed.Feed, name string, url string, err error) {
	feeds, err := s.loadFeeds()
	if err != nil {
		return nil, "", "", err
	}
	jsonFeed := feeds[key]
	if jsonFeed == nil {
		return nil, "", "", errors.New("feed not found")
	}
	return jsonFeed.Feed, jsonFeed.Name, jsonFeed.URL, nil
}

func (s *JsonStore) GetAllKeys() ([]string, error) {
	feeds, err := s.loadFeeds()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(feeds))
	for key := range feeds {
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *JsonStore) loadFeeds() (map[string]*JsonFeed, error) {
	file, err := os.Open(s.JsonFile)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]*JsonFeed), nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	var feeds map[string]*JsonFeed
	if err = jsonParser.Decode(&feeds); err != nil {
		return nil, err
	}
	return feeds, nil
}

func (s *JsonStore) saveFeeds(feeds map[string]*JsonFeed) error {
	file, err := os.Create(s.JsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonEncoder := json.NewEncoder(file)
	jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(feeds)
}
