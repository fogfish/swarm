package qtest

import (
	"log/slog"
	"os"
)

// Config logger for debug/testing purposes
func NewLogger() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		),
	)
}
