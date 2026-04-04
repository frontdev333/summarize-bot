package main

import (
	"frontdev333/summarize-bot/internal/config"
	"log/slog"
	"os"
	"time"

	"gopkg.in/telebot.v4"
)

func main() {
	token, err := config.LoadToken()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	settings := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(settings)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	b.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Send("Привет! Я готов. Используй /ping для проверки.")
	})

	b.Handle("/ping", func(ctx telebot.Context) error {
		return ctx.Send("pong")
	})

	b.Start()
}
