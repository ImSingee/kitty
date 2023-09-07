package glob

import (
	"path/filepath"
	"strings"

	"github.com/ImSingee/go-ex/glob"
)

func Match(pattern, name string) bool {
	if strings.Contains(pattern, "/") {
		return glob.Match(pattern, name)
	} else {
		return glob.Match(pattern, filepath.Base(name))
	}
}
