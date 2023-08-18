package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

func NewProvider() *slog.Logger {
	opts := slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true, // Add the line this code happened.
	}

	textHandler := opts.NewTextHandler(os.Stdout)
	logger := slog.New(textHandler)

	return logger
}
