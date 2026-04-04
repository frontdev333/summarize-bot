package config

import (
	"fmt"
	"os"
)

func LoadToken() (string, error) {
	token, ok := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if ok && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("you must to set env variable TELEGRAM_BOT_TOKEN")
}
