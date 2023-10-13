package lintstaged

import (
	"strings"

	"github.com/ImSingee/go-ex/ee"

	"github.com/ImSingee/kitty/internal/lib/ignore"
)

type IgnoreChecker struct {
	matcher ignore.Matcher
}

var ignoreFilenames = []string{
	".kittyignore",
	".lintstagedignore",
}

func NewIgnoreChecker(repoRoot string) (*IgnoreChecker, error) {
	ps, err := ignore.ReadPatterns(repoRoot, nil, ignoreFilenames...)
	if err != nil {
		return nil, ee.Wrap(err, "cannot read ignore patterns")
	}
	return &IgnoreChecker{ignore.NewMatcher(ps)}, nil
}

func (c *IgnoreChecker) ShouldIgnore(gitRelativePath string) bool {
	parts := strings.Split(gitRelativePath, "/")

	return c.matcher.Match(parts, false)
}
