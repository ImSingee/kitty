package glob

import (
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func Match(pattern string, g glob.Glob, name string) bool {
	if strings.Contains(pattern, "/") {
		return match(g, name)
	} else {
		return match(g, filepath.Base(name))
	}
}

func match(g glob.Glob, target string) bool {
	return g.Match(target)
}
