package summary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Summarizer interface {
	Summarize(text string, maxLen int) (string, error)
}

type FallbackSummarizer struct {
	primary   *GeminiClient
	secondary *Stub
}

func (f *FallbackSummarizer) Summarize(text string, maxLen int) (string, error) {
	if f.primary.key != "" {
		res, err := f.primary.Summarize(text, maxLen)
		if err == nil {
			return f.secondary.Summarize(res, maxLen)
		}
	}
	return f.secondary.Summarize(text, maxLen)
}

func NewFallbackSummarizer(model, key string) *FallbackSummarizer {
	return &FallbackSummarizer{
		primary: &GeminiClient{
			model: model,
			key:   key,
		},
		secondary: &Stub{},
	}
}

type Stub struct {
}

type GeminiClient struct {
	model string
	key   string
}

func (g *GeminiClient) Summarize(text string, maxLen int) (string, error) {
	prompt := fmt.Sprintf("Будь добр, сделай пожалуйста краткое резюме-выжимку на русском языке из текста на %d символов. Заверши выжимку полным предложением. Вот текст: %s", maxLen, text)

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model, g.key,
	)
	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": prompt}}},
		},
	}

	dto := &struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Data string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		ResponseId string `json:"response_id"`
	}{}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("failed to encode body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	c := http.Client{Timeout: 30 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(dto); err != nil {
		return "", err
	}

	fmt.Printf("%#v", dto)
	text = dto.Candidates[0].Content.Parts[0].Data

	return text, nil
}

func (s *Stub) Summarize(text string, maxLen int) (string, error) {
	runes := []rune(text)

	if len(runes) <= maxLen {
		return text, nil
	}

	runes = runes[:maxLen]

	for i := len(runes) - 1; i > 0; i-- {
		v := runes[i]
		if v == '.' || v == '!' || v == '?' {
			return string(runes[:i+1]), nil
		}
	}

	for i := len(runes) - 1; i > 0; i-- {
		v := runes[i]
		if v == ' ' {
			return string(runes[:i]) + "...", nil
		}
	}

	return string(runes[:maxLen]) + "...", nil
}
