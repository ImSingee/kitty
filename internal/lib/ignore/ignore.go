package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Matcher = gitignore.Matcher
type Pattern = gitignore.Pattern

const (
	commentPrefix = "#"
	gitDir        = ".git"
)

func NewMatcher(ps []Pattern) Matcher {
	return gitignore.NewMatcher(ps)
}

// ReadPatterns read and parse ignoreFileNames recursively
//
// The result is in the ascending order of priority (last higher).
func ReadPatterns(root string, dirs []string, ignoreFileNames ...string) ([]Pattern, error) {
	ps, err := readIgnoreFiles(root, dirs, ignoreFileNames...)
	if err != nil {
		return nil, err
	}

	sub, err := os.ReadDir(filepath.Join(root, filepath.Join(dirs...)))
	if err != nil {
		return nil, err
	}

	for _, fi := range sub {
		if !fi.IsDir() || fi.Name() == gitDir {
			continue
		}

		nextDirs := make([]string, 0, len(dirs)+1)
		nextDirs = append(nextDirs, dirs...)
		nextDirs = append(nextDirs, fi.Name())

		subps, err := ReadPatterns(root, nextDirs, ignoreFileNames...)
		if err != nil {
			return nil, err
		}
		ps = append(ps, subps...)
	}

	return ps, nil
}

func readIgnoreFiles(root string, dirs []string, ignoreFiles ...string) (ps []Pattern, err error) {
	for _, ignoreFile := range ignoreFiles {
		subps, err := readIgnoreFile(root, dirs, ignoreFile)
		if err != nil {
			return nil, err
		}

		ps = append(ps, subps...)
	}
	return
}

// readIgnoreFile reads a specific git ignore file.
func readIgnoreFile(root string, dirs []string, ignoreFile string) (ps []Pattern, err error) {
	filename := filepath.Join(root, filepath.Join(dirs...), ignoreFile)

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(s, commentPrefix) {
			continue
		}

		ps = append(ps, gitignore.ParsePattern(s, dirs))
	}

	return
}
