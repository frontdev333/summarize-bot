package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramBotToken string
	LogLevel         string
}

func LoadConfig() (Config, error) {
	token, ok := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !ok || token == "" {
		return Config{}, fmt.Errorf("you must to set env variable TELEGRAM_BOT_TOKEN")
	}

	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok || level == "" {
		level = "info"
	}

	return Config{
		TelegramBotToken: token,
		LogLevel:         level,
	}, nil
}
