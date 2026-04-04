package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	BotToken string `yaml:"bot_token"`
}

func LoadToken() (string, error) {
	token, ok := os.LookupEnv("BOT_TOKEN")
	if ok && token != "" {
		return token, nil
	}

	file, err := os.ReadFile("config.yaml")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("read config.yaml file: %w", err)
			return "", err
		}

		token, err = inputToken()
		if err != nil {
			return "", err
		}

		return token, nil
	}

	cfg := Config{}

	if err = yaml.Unmarshal(file, &cfg); err != nil {
		return "", fmt.Errorf("unmarshal config.yaml: %w", err)
	}

	if cfg.BotToken != "" {
		return cfg.BotToken, nil
	}

	token, err = inputToken()
	if err != nil {
		return "", err
	}

	return token, nil
}

func inputToken() (string, error) {
	fmt.Println("Введите пожалуйста токен бота: ")
	var token string
	if _, err := fmt.Scan(&token); err != nil {
		return "", fmt.Errorf("fmt.Scan read token: %w", err)
	}

	token = strings.TrimSpace(token)

	if token == "" {
		fmt.Println("токен не может быть пустым!")
		return "", fmt.Errorf("user input is empty")
	}

	return token, nil
}
