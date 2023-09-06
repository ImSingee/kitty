package xlog

import (
	"context"
	"log/slog"
)

var DisabledLogger = slog.New(DisabledLogHandler{})

type DisabledLogHandler struct{}

func (d DisabledLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (d DisabledLogHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (d DisabledLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return d
}

func (d DisabledLogHandler) WithGroup(name string) slog.Handler {
	return d
}
