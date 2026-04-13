package config

import (
	"briefpush/notify"
	"encoding/json"
	"os"
)

type FeedConfig struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	BaseURL      string        `json:"base_url"`
	ApiKey       string        `json:"api_key"`
	Model        string        `json:"model"`
	JsonFilePath string        `json:"json_file_path"`
	ReportHour   uint          `json:"report_hour"`
	Feeds        []FeedConfig  `json:"feeds"`
	Notification notify.Config `json:"notification"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
