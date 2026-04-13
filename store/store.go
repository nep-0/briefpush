package store

import "github.com/mmcdole/gofeed"

type FeedStorer interface {
	AddFeed(key string, name string, url string, feed *gofeed.Feed) error
	UpdateFeed(key string, feed *gofeed.Feed) error
	GetFeed(key string) (feed *gofeed.Feed, name string, url string, err error)
	GetAllKeys() ([]string, error)
}
