package feed

import (
	"briefpush/ai"
	"briefpush/config"
	"briefpush/notify"
	"briefpush/store"
	"context"
	"log"
	"time"

	"github.com/sourcegraph/conc/iter"

	"github.com/go-co-op/gocron/v2"
	"github.com/mmcdole/gofeed"
)

type FeedManager struct {
	Parser    *gofeed.Parser
	Store     store.FeedStorer
	AI        ai.AI
	Notify    *notify.Dispatcher
	Scheduler gocron.Scheduler
}

func NewFeedManager(store store.FeedStorer, ai ai.AI, notify *notify.Dispatcher, reportHour uint, feeds []config.FeedConfig) *FeedManager {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	fm := &FeedManager{
		Parser:    gofeed.NewParser(),
		Store:     store,
		AI:        ai,
		Notify:    notify,
		Scheduler: s,
	}

	for _, feed := range feeds {
		err := fm.AddFeedIfNotExists(feed.Key, feed.Name, feed.URL)
		if err != nil {
			log.Printf("Error adding feed %s: %v", feed.Key, err)
		}
	}

	_, err = fm.Scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(reportHour, 0, 0))),
		gocron.NewTask(
			func() {
				ctx := context.Background()
				_, err := fm.SendReports(ctx, 24*time.Hour)
				if err != nil {
					log.Printf("Error sending reports: %v", err)
				}
			},
		),
	)
	_, err = fm.Scheduler.NewJob(
		gocron.DurationJob(30*time.Minute),
		gocron.NewTask(
			func() {
				err := fm.UpdateFeeds()
				if err != nil {
					log.Printf("Error updating feeds: %v", err)
				}
			},
		),
	)
	return fm
}

func (m *FeedManager) Start() {
	m.Scheduler.Start()
}

func (m *FeedManager) addFeed(key string, name string, url string, feed *gofeed.Feed) error {
	return m.Store.AddFeed(key, name, url, feed)
}

func (m *FeedManager) AddFeedIfNotExists(key string, name string, url string) error {
	_, _, _, err := m.Store.GetFeed(key)
	if err == nil {
		return nil
	}
	parsedFeed, err := m.Parser.ParseURL(url)
	if err != nil {
		return err
	}
	return m.addFeed(key, name, url, parsedFeed)
}

func (m *FeedManager) UpdateFeed(key string) error {
	log.Printf("Updating feed %s", key)
	_, _, url, err := m.Store.GetFeed(key)
	if err != nil {
		return err
	}
	parsedFeed, err := m.Parser.ParseURL(url)
	if err != nil {
		return err
	}
	return m.Store.UpdateFeed(key, parsedFeed)
}

func (m *FeedManager) UpdateFeeds() error {
	keys, err := m.Store.GetAllKeys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		err = m.UpdateFeed(key)
		if err != nil {
			log.Printf("Error updating feed %s: %v", key, err)
		} else {
			log.Printf("Successfully updated feed %s", key)
		}
	}
	return nil
}

func (m *FeedManager) GetFeed(key string) (*gofeed.Feed, string, string, error) {
	return m.Store.GetFeed(key)
}

func (m *FeedManager) GetAllFeeds() (map[string]*store.JsonFeed, error) {
	keys, err := m.Store.GetAllKeys()
	if err != nil {
		return nil, err
	}
	feeds := make(map[string]*store.JsonFeed)
	for _, key := range keys {
		feed, _, _, err := m.Store.GetFeed(key)
		if err != nil {
			return nil, err
		}
		feeds[key] = &store.JsonFeed{Feed: feed}
	}
	return feeds, nil
}

func (m *FeedManager) GenerateReport(ctx context.Context, key string, duration time.Duration) (string, error) {
	feed, _, _, err := m.Store.GetFeed(key)
	if err != nil {
		return "", err
	}
	summaries := []string{}
	total := len(feed.Items)
	log.Printf("Generating report for feed %s with %d items", key, total)
	iter.Map(feed.Items, func(item **gofeed.Item) error {
		if time.Since(*(*item).PublishedParsed) > duration {
			return nil
		}
		summary, err := m.AI.GenerateSummary(ctx, (*item).Title, (*item).Description)
		if err != nil {
			log.Printf("Error generating summary for item in feed %s: %v", key, err)
			return err
		}
		log.Printf("Generated summary for item %d/%d: %s", len(summaries)+1, total, (*item).Title)
		summaries = append(summaries, summary)
		return nil
	})
	return m.AI.GenerateReport(ctx, summaries)
}

func (m *FeedManager) GenerateReports(ctx context.Context, duration time.Duration) (map[string]string, error) {
	keys, err := m.Store.GetAllKeys()
	if err != nil {
		return nil, err
	}
	reports := make(map[string]string)
	total := len(keys)
	log.Printf("Generating reports for %d feeds", total)
	for i, key := range keys {
		report, err := m.GenerateReport(ctx, key, duration)
		if err != nil {
			log.Printf("Error generating report for feed %s: %v", key, err)
			continue
		}
		log.Printf("Generated report for feed %d/%d: %s", i+1, total, key)
		reports[key] = report
	}
	return reports, nil
}

func (m *FeedManager) SendReport(ctx context.Context, key string, duration time.Duration) (string, error) {
	report, err := m.GenerateReport(ctx, key, duration)
	if err != nil {
		return "", err
	}
	err = m.Notify.Notify(ctx, notify.Message{
		Subject: "Daily Report for " + key,
		Body:    report,
	})
	if err != nil {
		return "", err
	}
	return report, nil
}

func (m *FeedManager) SendReports(ctx context.Context, duration time.Duration) (map[string]string, error) {
	reports, err := m.GenerateReports(ctx, duration)
	if err != nil {
		return nil, err
	}
	for key, report := range reports {
		err = m.Notify.Notify(ctx, notify.Message{
			Subject: "Daily Report for " + key,
			Body:    report,
		})
		if err != nil {
			return nil, err
		}
	}
	return reports, nil
}
