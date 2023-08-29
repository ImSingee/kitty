package lintstaged

import "log/slog"

func debug() bool {
	return slog.Default().Enabled(nil, slog.LevelDebug)
}
