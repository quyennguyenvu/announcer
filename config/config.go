package config

import "os"

type Config struct {
	BreakfastConfig *BreakfastConfig
}

type BreakfastConfig struct {
	BreakfastLink       string
	DiscordWebhookURL   string
	BlessingMessageLink string
	BlessingUpdateLink  string
}

// load configuration settings from env
func LoadConfig() *Config {
	return &Config{
		BreakfastConfig: &BreakfastConfig{
			BreakfastLink:       os.Getenv("BREAKFAST_LINK"),
			DiscordWebhookURL:   os.Getenv("DISCORD_WEBHOOK_URL"),
			BlessingMessageLink: os.Getenv("BLESSING_MESSAGE_LINK"),
			BlessingUpdateLink:  os.Getenv("BLESSING_UPDATE_LINK"),
		},
	}
}
