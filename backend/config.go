package main

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type config struct {
	App struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"app"`

	Google struct {
		User           string   `yaml:"user"`
		Password       string   `yaml:"password"`
		ClientID       string   `yaml:"client_id"`
		ClientSecret   string   `yaml:"client_secret"`
		RedirectURL    string   `yaml:"redirect_url"`
		Scopes         []string `yaml:"scopes"`
		ProjectID      string   `yaml:"project_id"`
		TopicID        string   `yaml:"topic_id"`
		SubscriptionID string   `yaml:"subscription_id"`
		CredentialsJSON string  `yaml:"credentials_json"`
	} `yaml:"google"`
}

func loadConfig() (*config, error) {
	var cfg config

	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("mở file config thất bại: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("đọc file config thất bại: %w", err)
	}

	return &cfg, nil
}
