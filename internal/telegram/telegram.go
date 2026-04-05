package telegram

import (
	"time"

	"gopkg.in/telebot.v4"
)

type Bot interface {
	Start() error
	Stop() error
}

type MinimalBot struct {
	tbot *telebot.Bot
}

func NewMinimalBot(token string) (Bot, error) {
	settings := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}

	return &MinimalBot{tbot: bot}, nil
}

func (b *MinimalBot) Start() error {
	b.tbot.Start()
	return nil
}

func (b *MinimalBot) Stop() error {
	b.tbot.Stop()
	return nil
}

func (b *MinimalBot) Underlying() *telebot.Bot {
	return b.tbot
}
