package config

import "os"

type Config struct {
	BreakfastLink     string
	DiscordWebhookURL string
}

// load configuration settings from env
func LoadConfig() *Config {
	return &Config{
		BreakfastLink:     os.Getenv("BREAKFAST_LINK"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}
}
