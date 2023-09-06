package hooks

import (
	"fmt"
	"os"
	"strings"
)

func l(msg string, args ...any) {
	s := msg
	if len(args) != 0 {
		s = fmt.Sprintf(msg, args...)
	}

	_, _ = os.Stderr.Write([]byte("kitty - " + strings.TrimSpace(s) + "\n"))
}
