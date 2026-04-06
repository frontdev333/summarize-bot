package news

import (
	"encoding/json"
	"fmt"
	"frontdev333/summarize-bot/internal/subscriptions"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

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

type FallbackProvider struct {
	primary   Provider
	secondary Provider
}

func (p *FallbackProvider) FetchByTopic(topic string, limit int) ([]Article, error) {
	articles, err := p.primary.FetchByTopic(topic, limit)
	if err == nil {
		return articles, nil
	}
	slog.Error("primary provider", "error", err)

	return p.secondary.FetchByTopic(topic, limit)
}

type MockProvider struct {
	Articles map[string][]Article
}

type NewsAPIClient struct {
	client  http.Client
	apiKey  string
	baseUrl string
}

func (p *NewsAPIClient) FetchByTopic(topic string, limit int) ([]Article, error) {
	req, err := http.NewRequest(http.MethodGet, p.baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", p.apiKey)
	params := url.Values{}
	params.Set("q", topic)
	req.URL.RawQuery = params.Encode()

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var articles NewsAPIArticles
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status", "http_status", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, err
	}

	res := make([]Article, articles.TotalResults)

	i := 0
	for _, a := range articles.Articles {
		if i >= limit {
			break
		}

		res[i] = Article{
			ID:          "1",
			Title:       a.Title,
			URL:         a.Url,
			Source:      a.Source.Name,
			Description: a.Description,
		}
		i++
	}

	return res, nil
}

func NewNewsAPIClient(apiKey, baseURL string) *NewsAPIClient {
	return &NewsAPIClient{
		client:  http.Client{Timeout: 5 * time.Second},
		apiKey:  apiKey,
		baseUrl: baseURL,
	}
}

func NewFallbackProvider(primary, secondary Provider) *FallbackProvider {
	return &FallbackProvider{
		primary:   primary,
		secondary: secondary,
	}
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

		limit := 10
		for _, v := range topics {
			partArticles, err := prov.FetchByTopic(v, limit)
			if err != nil {
				slog.Error("can not fetch articles", "topic", v, "error", err)
				continue
			}

			allArticles = append(allArticles, partArticles...)
		}

		var res strings.Builder
		for _, a := range allArticles {
			if limit == 0 {
				break
			}
			articleCard := fmt.Sprintf("\nЗаголовок: %s\nСсылка: %s\nИсточник: %s\n", a.Title, a.URL, a.Source)
			res.WriteString(articleCard)
			limit--
		}

		fmt.Println(res.String())

		return ctx.Send(res.String())
	})
}
