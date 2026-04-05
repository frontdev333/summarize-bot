package main

import (
	"frontdev333/summarize-bot/internal/config"
	"frontdev333/summarize-bot/internal/telegram"
	"log/slog"
	"os"

	"gopkg.in/telebot.v4"
)

func main() {
	token, err := config.LoadToken()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	b, err := telegram.NewMinimalBot(token)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	tgbot, ok := b.(*telegram.MinimalBot)
	if !ok {
		slog.Error("failed to cast bot to MinimalBot")
		os.Exit(1)
	}

	tgbot.Underlying().Handle("/start", func(ctx telebot.Context) error {
		return ctx.Send("Привет! Я готов. Используй /ping для проверки.")
	})

	tgbot.Underlying().Handle("/ping", func(ctx telebot.Context) error {
		return ctx.Send("pong")
	})

	slog.Info("bot started, polling updates")
	b.Start()
}
