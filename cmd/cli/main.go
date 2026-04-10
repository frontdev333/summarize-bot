package main

import (
	"frontdev333/summarize-bot/internal/app"
	"log/slog"
	"os"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("app.Run", "error", err)
		os.Exit(1)
	}
}
