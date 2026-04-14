package news

import (
	"encoding/json"
	"errors"
	"fmt"
	"frontdev333/summarize-bot/internal/cache"
	"frontdev333/summarize-bot/internal/subscriptions"
	"frontdev333/summarize-bot/internal/summary"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/telebot.v4"
)

var ProviderResponseError = errors.New("provider is unavailable")

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
		return nil, fmt.Errorf("http status: %d %w", resp.StatusCode, ProviderResponseError)
	}

	if err = json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, err
	}

	res := make([]Article, 0, limit)

	i := 0
	for _, a := range articles.Articles {
		if i >= limit {
			break
		}

		res[i] = Article{
			ID:          a.Url,
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

func RegisterNewsHandlers(
	b *telebot.Bot,
	subs subscriptions.SubscriptionStore,
	prov Provider,
	summarizer summary.Summarizer,
	cash *cache.SummaryCache,
	maxNewsPerTopic,
	maxNewsPerReq,
	maxParallel int,
) {

	b.Handle("/news", func(ctx telebot.Context) error {
		topics := subs.GetTopics(ctx.Sender().ID)
		if len(topics) == 0 {
			return ctx.Send("У вас нет подписок. Добавьте темы через /add <topic>")
		}

		allArticles, err := getAllArticles(topics, prov, maxNewsPerTopic)
		if err != nil {
			return ctx.Send("Новости временно недоступны, попробуйте позже")
		}

		res, err := makeArticlesMessage(allArticles, maxNewsPerReq, cash, summarizer, maxParallel)
		if err != nil {
			return err
		}

		return ctx.Send(res)
	})
}

func makeArticlesMessage(
	allArticles []Article,
	limit int,
	cash *cache.SummaryCache,
	summarizer summary.Summarizer,
	maxParallel int,
) (string, error) {
	var res strings.Builder
	slog.Info("articles fetched", "limit_exceeded_count", len(allArticles)-limit)
	descriptions := SummarizeInParallelSimple(allArticles, summarizer, cash, maxParallel)
	for i, a := range allArticles {
		if limit == 0 {
			break
		}

		articleCard := fmt.Sprintf("\nЗаголовок: %s\nОписание: %s\nСсылка: %s\nИсточник: %s\n", a.Title, descriptions[i], a.URL, a.Source)
		res.WriteString(articleCard)
		limit--
	}
	return res.String(), nil
}

func getAllArticles(topics []string, prov Provider, MaxNewsPerTopic int) ([]Article, error) {
	allArticles := make([]Article, 0)

	for _, v := range topics {
		partArticles, err := prov.FetchByTopic(v, MaxNewsPerTopic)
		if err != nil {
			if errors.Is(err, ProviderResponseError) {
				return nil, err
			}
			slog.Error("can not fetch articles", "topic", v, "error", err)
			continue
		}

		allArticles = append(allArticles, partArticles...)
	}
	return deduplicateArticles(allArticles), nil
}

func deduplicateArticles(articles []Article) []Article {
	seen := make(map[string]struct{})
	var unique []Article

	duplicateCounter := 0

	for _, a := range articles {
		normalizedUrl, err := normalizeUrl(a)
		if err != nil {
			slog.Error("parse url", "error", err)
			continue
		}

		if _, ok := seen[normalizedUrl]; ok {
			duplicateCounter++
			continue
		}
		seen[normalizedUrl] = struct{}{}
		unique = append(unique, a)
	}

	slog.Info("duplicates", "count", duplicateCounter)

	return unique
}

func normalizeUrl(a Article) (string, error) {
	parsedUrl, err := url.Parse(a.ID)
	if err != nil {
		return "", err
	}

	parsedUrl.RawQuery = ""
	parsedUrl.Fragment = ""

	return parsedUrl.String(), nil
}

func SummarizeInParallelSimple(
	articles []Article,
	summarizer summary.Summarizer,
	cache *cache.SummaryCache,
	maxParallel int,
) []string {
	wg := &sync.WaitGroup{}
	results := make([]string, len(articles))
	sem := make(chan struct{}, maxParallel)
	var failedCounter atomic.Uint32
	var successCounter atomic.Uint32

	for i, a := range articles {
		cachedDesc, ok := cache.Get(a.ID)
		if ok {
			results[i] = cachedDesc
			continue
		}

		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() {
				<-sem
			}()

			desc, err := summarizer.Summarize(a.Description, 255)
			if err != nil {
				slog.Error("parallel summarize", "error", err)
				failedCounter.Add(1)
				return
			}
			cache.Set(a.ID, desc)
			results[id] = desc
			successCounter.Add(1)
		}(i)
	}

	wg.Wait()
	slog.Info("Parallel summarize is done", "success_count", successCounter.Load(), "fails_count", failedCounter.Load())

	return results
}
