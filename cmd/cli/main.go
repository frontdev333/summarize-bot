package main

import (
	"frontdev333/summarize-bot/internal/config"
	"frontdev333/summarize-bot/internal/subscriptions"
	"frontdev333/summarize-bot/internal/telegram"
	"log/slog"
	"os"
	"os/signal"
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

	store, err := subscriptions.NewFileStore(cfg.StorePath)
	if err != nil {
		slog.Error("create file store", "error", err)
		os.Exit(1)
	}

	telegram.RegisterCoreHandlers(tgbot.Underlying())
	telegram.RegisterSubscriptionHandlers(tgbot.Underlying(), store, 64)

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
