package main

import (
	"frontdev333/summarize-bot/internal/config"
	"frontdev333/summarize-bot/internal/telegram"
	"log/slog"
	"os"
	"os/signal"

	"gopkg.in/telebot.v4"
)

func main() {
	shtdwn := make(chan os.Signal)
	signal.Notify(shtdwn, os.Interrupt)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	b, err := telegram.NewMinimalBot(cfg.TelegramBotToken)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	tgbot, ok := b.(*telegram.MinimalBot)
	if !ok {
		slog.Error("failed to cast bot to MinimalBot")
		os.Exit(1)
	}

	RegisterCoreHandlers(tgbot.Underlying())

	go func() {
		<-shtdwn
		slog.Info("bot stopping")

		if err = b.Stop(); err != nil {
			slog.Error("bot stop: %w", err)
			return
		}

		slog.Info("bot stopped")
	}()

	slog.Info("bot started, polling updates")
	b.Start()
}

func RegisterCoreHandlers(b *telebot.Bot) {
	b.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Send("Привет! Я готов. Используй /ping для проверки.")
	})

	b.Handle("/ping", func(ctx telebot.Context) error {
		return ctx.Send("pong")
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		return ctx.Send("Неизвестная команда. Доступно: /start, /ping")
	})
}
