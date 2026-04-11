package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken     string
	LogLevel             string
	StorePath            string
	NewsAPIKey           string
	GeminiAPIKey         string
	GeminiModel          string
	MaxTopics            int
	MaxNewsPerReq        int
	MaxNewsPerTopic      int
	GeminiMaxConcurrency int
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

	var err error

	var maxTopics int
	maxTopicsEnv, ok := os.LookupEnv("MAX_TOPICS_PER_USER")
	if !ok {
		maxTopics = 10
	} else {
		maxTopics, err = strconv.Atoi(maxTopicsEnv)
		if err != nil {
			return Config{}, err
		}
	}

	var maxNewsPerReq int
	maxNewsPerReqEnv, ok := os.LookupEnv("MAX_NEWS_PER_REQUEST")
	if !ok {
		maxNewsPerReq = 10
	} else {
		var err error
		maxNewsPerReq, err = strconv.Atoi(maxNewsPerReqEnv)
		if err != nil {
			return Config{}, err
		}
	}

	var maxNewsPerTopic int
	maxNewsPerTopicEnv, ok := os.LookupEnv("MAX_TOPICS_PER_USER")
	if !ok {
		maxNewsPerTopic = 0
	} else {
		var err error
		maxNewsPerTopic, err = strconv.Atoi(maxNewsPerTopicEnv)
		if err != nil {
			return Config{}, err
		}
	}

	var geminiMaxConcurrency int
	geminiMaxConcurrencyEnv, ok := os.LookupEnv("GEMINI_MAX_CONCURRENCY")
	if !ok {
		geminiMaxConcurrency = 4
	} else {
		var err error
		geminiMaxConcurrency, err = strconv.Atoi(geminiMaxConcurrencyEnv)
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		TelegramBotToken:     token,
		LogLevel:             level,
		StorePath:            path,
		NewsAPIKey:           key,
		GeminiAPIKey:         apiKey,
		GeminiModel:          model,
		MaxTopics:            maxTopics,
		MaxNewsPerReq:        maxNewsPerReq,
		MaxNewsPerTopic:      maxNewsPerTopic,
		GeminiMaxConcurrency: geminiMaxConcurrency,
	}, nil
}
