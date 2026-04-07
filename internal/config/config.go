package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramBotToken string
	LogLevel         string
	StorePath        string
	NewsAPIKey       string
	GeminiAPIKey     string
	GeminiModel      string
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

	path, ok := os.LookupEnv("STORE_PATH")
	if !ok || path == "" {
		path = "subscriptions.json"
	}

	key, ok := os.LookupEnv("NEWSAPI_KEY")
	if !ok {
		key = ""
	}

	apiKey, ok := os.LookupEnv("GOOGLE_API_KEY")
	if !ok {
		apiKey = ""
	}

	model, ok := os.LookupEnv("GEMINI_MODEL")
	if !ok {
		model = "gemini-2.5-flash"
	}

	return Config{
		TelegramBotToken: token,
		LogLevel:         level,
		StorePath:        path,
		NewsAPIKey:       key,
		GeminiAPIKey:     apiKey,
		GeminiModel:      model,
	}, nil
}
