package telegram

import (
	"fmt"
	"frontdev333/summarize-bot/internal/subscriptions"
	"log/slog"
	"strings"
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

func RegisterSubscriptionHandlers(b *telebot.Bot, subs subscriptions.SubscriptionStore, maxTopics int) {
	b.Handle("/add", func(ctx telebot.Context) error {
		topics := subs.GetTopics(ctx.Sender().ID)
		if len(topics) >= maxTopics {
			return ctx.Send(fmt.Sprintf("Достигнут лимит тем (%d)", maxTopics))
		}

		topic, _ := strings.CutPrefix(ctx.Text(), "/add")
		topic = strings.TrimSpace(topic)

		if err := subs.AddTopic(ctx.Sender().ID, topic); err != nil {
			slog.Info("user try add topic", "userID", ctx.Sender().ID, "error", err)
			return ctx.Send("Вы уже подписаны на тему " + topic)
		}
		return ctx.Send("Вы подписаны на тему " + topic)
	})

	b.Handle("/remove", func(ctx telebot.Context) error {
		topic, _ := strings.CutPrefix(ctx.Text(), "/remove")
		topic = strings.TrimSpace(topic)
		if err := subs.RemoveTopic(ctx.Sender().ID, topic); err != nil {
			slog.Info("user try remove topic", "userID", ctx.Sender().ID, "error", err)
			return ctx.Send("Вы не подписаны на тему " + topic)
		}
		return ctx.Send("Вы отписались от темы " + topic)
	})

	b.Handle("/topics", func(ctx telebot.Context) error {
		topics := subs.GetTopics(ctx.Sender().ID)
		if len(topics) == 0 {
			return ctx.Send("У вас нет действующих подписок")
		}

		var res strings.Builder

		for _, v := range topics {
			res.WriteString("\n- " + v)
		}

		return ctx.Send("Ваши подписки: " + res.String())
	})
}
