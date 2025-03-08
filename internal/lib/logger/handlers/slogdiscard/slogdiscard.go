package slogdiscard

import (
	"context"
	"log/slog"
)

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

type DiscardHandler struct{}

func (d *DiscardHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (d *DiscardHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (d *DiscardHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return d
}

func (d *DiscardHandler) WithGroup(name string) slog.Handler {
	return d
}
