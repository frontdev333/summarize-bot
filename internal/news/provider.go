package news

import (
	"fmt"
	"frontdev333/summarize-bot/internal/subscriptions"
	"log/slog"
	"strings"

	"gopkg.in/telebot.v4"
)

type Provider interface {
	FetchByTopic(topic string, limit int) ([]Article, error)
}

type Article struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type MockProvider struct {
	Articles map[string][]Article
}

func (p *MockProvider) FetchByTopic(topic string, limit int) ([]Article, error) {
	arts := p.Articles[topic]
	if len(arts) > limit {
		arts = arts[:limit]
	}
	return arts, nil
}

func RegisterNewsHandlers(b *telebot.Bot, subs subscriptions.SubscriptionStore, prov Provider) {
	b.Handle("/news", func(ctx telebot.Context) error {
		topics := subs.GetTopics(ctx.Sender().ID)
		if len(topics) == 0 {
			return ctx.Send("У вас нет подписок. Добавьте темы через /add <topic>")
		}
		allArticles := make([]Article, 0)

		for _, v := range topics {
			partArticles, err := prov.FetchByTopic(v, 10)
			if err != nil {
				slog.Error("can not fetch articles", "topic", v, "error", err)
				continue
			}

			allArticles = append(allArticles, partArticles...)
		}

		var res strings.Builder
		for _, a := range allArticles {
			articleCard := fmt.Sprintf("\nЗаголовок: %s\nСсылка: %s\nИсточник: %s\n", a.Title, a.URL, a.Source)
			res.WriteString(articleCard)
		}

		return ctx.Send(res.String())
	})
}
