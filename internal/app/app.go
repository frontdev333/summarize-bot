package app

import (
	"fmt"
	"frontdev333/summarize-bot/internal/cache"
	"frontdev333/summarize-bot/internal/config"
	"frontdev333/summarize-bot/internal/news"
	"frontdev333/summarize-bot/internal/subscriptions"
	"frontdev333/summarize-bot/internal/summary"
	"frontdev333/summarize-bot/internal/telegram"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func Run() error {
	shtdwn := make(chan os.Signal)
	signal.Notify(shtdwn, os.Interrupt)

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	b, err := telegram.NewMinimalBot(cfg.TelegramBotToken)
	if err != nil {
		return err
	}

	tgbot, ok := b.(*telegram.MinimalBot)
	if !ok {
		return fmt.Errorf("failed to cast bot to MinimalBot: %w", err)
	}

	store, err := subscriptions.NewFileStore(cfg.StorePath)
	if err != nil {
		return fmt.Errorf("create file store: %w", err)
	}

	primary := news.NewNewsAPIClient(cfg.NewsAPIKey, "https://newsapi.org/v2/everything")

	secondary := &news.MockProvider{
		Articles: map[string][]news.Article{
			"golang": {
				{Title: "Go 1.24 Released", Source: "go.dev", URL: "https://go.dev/blog/"},
				{Title: "Understanding Goroutines", Source: "go.dev", URL: "https://go.dev/doc/"},
			},
			"backend": {
				{Title: "Designing Reliable APIs", Source: "example.com", URL: "https://example.com/reliable-apis"},
				{Title: "Caching Strategies in Microservices", Source: "example.com", URL: "https://example.com/caching-microservices"},
			},
		},
	}

	var provider news.Provider

	if cfg.NewsAPIKey != "" {
		provider = news.NewFallbackProvider(primary, secondary)
	} else {
		provider = secondary
	}

	summarizer := summary.NewFallbackSummarizer(cfg.GeminiModel, cfg.GeminiAPIKey)

	cash := cache.NewSummaryCache(10 * time.Minute)

	news.RegisterNewsHandlers(tgbot.Underlying(), store, provider, summarizer, cash, cfg.MaxNewsPerTopic, cfg.MaxNewsPerReq)
	telegram.RegisterCoreHandlers(tgbot.Underlying())
	telegram.RegisterSubscriptionHandlers(tgbot.Underlying(), store, cfg.MaxTopics)

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

	return nil
}
