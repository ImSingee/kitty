package lintstaged

import (
	"fmt"
	"log/slog"
	"os"
)

func unsetEnv(name string) {
	if v := os.Getenv(name); v != "" {
		slog.Debug(fmt.Sprintf("Unset %s (was `%s`)", name, v))
		os.Unsetenv(name)
	}
}
