package lintstaged

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exstrings"
	"github.com/ImSingee/go-ex/mr"

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
	ps, err := parseIgnoreRules(repoRoot)
	if err != nil {
		return nil, ee.Wrap(err, "cannot read ignore patterns")
	}

	return &IgnoreChecker{ignore.NewMatcher(ps)}, nil
}

func (c *IgnoreChecker) ShouldIgnore(gitRelativePath string) bool {
	parts := strings.Split(gitRelativePath, "/")

	return c.matcher.Match(parts, false)
}

func parseIgnoreRules(gitDir string) ([]ignore.Pattern, error) {
	cachedFiles, err := getCachedFiles(gitDir)
	if err != nil {
		return nil, ee.Wrap(err, "cannot get list of known files")
	}

	possibleIgnoreRuleFiles := cachedFiles

	possibleIgnoreRuleFiles = mr.Filter(possibleIgnoreRuleFiles, func(file string, _index int) bool {
		return exstrings.InStringList(ignoreFilenames, filepath.Base(file))
	})

	slog.Debug("Found possible ignore rule files", "possibleIgnoreRuleFiles", possibleIgnoreRuleFiles, "possibleIgnoreRuleFilesCount", len(possibleIgnoreRuleFiles))

	ps, err := ignore.ParsePatternFiles(gitDir, possibleIgnoreRuleFiles...)

	slog.Debug("Load ignore patterns", "count", len(ps))

	return ps, err
}
