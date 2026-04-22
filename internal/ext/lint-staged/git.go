package lintstaged

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ImSingee/kitty/internal/lib/git"
)

func execGit(args []string, dir string) (string, error) {
	result := git.R(dir, append([]string{"-c", "submodule.recurse=false"}, args...))

	return strings.TrimSpace(string(result.Output)), result.Err()
}

func normalizePath(path string) string {
	return filepath.Clean(path)
}

func determineGitDir(cwd string, relativeDir string) string {
	// if relative dir and cwd have different endings normalize it
	// this happens under windows, where normalize is unable to normalize git's output
	relativeDir = strings.TrimSuffix(relativeDir, string(filepath.Separator))

	if relativeDir != "" {
		// the current working dir is inside the git top-level directory
		return normalizePath(cwd[:strings.LastIndex(cwd, relativeDir)])
	} else {
		// the current working dir is the top-level git directory
		return normalizePath(cwd)
	}
}

func resolveGitConfigDir(gitDir string) (string, error) {
	file := normalizePath(filepath.Join(gitDir, ".git"))
	fileInfo, err := os.Stat(file)

	if os.IsNotExist(err) {
		return "", err
	}

	if fileInfo.IsDir() {
		return file, nil
	}

	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	contents := string(b)

	return filepath.Clean(strings.TrimPrefix(contents, "gitdir: ")), nil
}

func resolveGitRepo(cwd string) (gitDir, gitConfigDir string, err error) {
	// Unset GIT_DIR before running any git operations in case it's pointing to an incorrect location
	unsetEnv("GIT_DIR")
	unsetEnv("GIT_WORK_TREE")

	// read the path of the current directory relative to the top-level directory
	// don't read the toplevel directly, it will lead to an posix conform path on non posix systems (cygwin)
	gitRel, err := execGit([]string{"rev-parse", "--show-prefix"}, cwd)
	if err != nil {
		return "", "", err
	}

	gitDir = determineGitDir(normalizePath(cwd), gitRel)

	gitConfigDir, err = resolveGitConfigDir(gitDir)
	if err != nil {
		return "", "", err
	}

	slog.Debug("Resolved git directory", "gitDir", gitDir)
	slog.Debug("Resolved git config directory", "gitConfigDir", gitConfigDir)

	return gitDir, gitConfigDir, nil
}

func getDiffCommand(diff, diffFilter string) []string {
	diffFilter = normalizeDiffFilter(diffFilter)

	diffArgs := strings.Fields(strings.TrimSpace(diff))
	if len(diffArgs) == 0 {
		// Fall back to staged files when no explicit diff range is provided.
		diffArgs = []string{"--staged"}
	}

	// Docs for -z option:
	// https://git-scm.com/docs/git-diff#Documentation/git-diff.txt--z
	return append([]string{"diff", "--name-only", "-z", "--diff-filter=" + diffFilter}, diffArgs...)
}

func normalizeDiffFilter(diffFilter string) string {
	if diffFilter == "" {
		// Docs for --diff-filter option:
		// https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---diff-filterACDMRTUXB82308203
		diffFilter = "ACMR"
	}

	return diffFilter
}

func getStatusDiffCommand(diffFilter string, staged bool) []string {
	args := []string{"diff", "--name-only", "-z", "--diff-filter=" + normalizeDiffFilter(diffFilter)}
	if staged {
		args = append(args, "--staged")
	}

	// Docs for -z option:
	// https://git-scm.com/docs/git-diff#Documentation/git-diff.txt--z
	return args
}

func getSelectedFiles(options *Options, gitDir string) ([]string, error) {
	if options.Diff != "" {
		return execGitZ(getDiffCommand(options.Diff, options.DiffFilter), gitDir)
	}

	switch options.SelectionMode() {
	case SelectionModeStaged:
		return getStagedFiles(options.DiffFilter, gitDir)
	case SelectionModeUnstaged:
		return getUnstagedFiles(options.DiffFilter, gitDir)
	case SelectionModeUntracked:
		return getUntrackedFiles(gitDir)
	case SelectionModeTracked:
		return getTrackedFiles(options.DiffFilter, gitDir)
	case SelectionModeChanged:
		return getChangedFiles(options.DiffFilter, gitDir)
	case SelectionModeAll:
		tracked, err := getCachedFiles(gitDir)
		if err != nil {
			return nil, err
		}
		untracked, err := getUntrackedFiles(gitDir)
		if err != nil {
			return nil, err
		}

		return uniqueStrings(append(tracked, untracked...)), nil
	default:
		return nil, fmt.Errorf("unsupported selection mode %q", options.SelectionMode())
	}
}

// getStagedFiles returns a list of staged files in relative path to git root
func getStagedFiles(diffFilter string, gitDir string) ([]string, error) {
	lines, err := execGitZ(getStatusDiffCommand(diffFilter, true), gitDir)
	if err != nil {
		return nil, err
	}

	return lines, nil
}

func getUnstagedFiles(diffFilter string, gitDir string) ([]string, error) {
	lines, err := execGitZ(getStatusDiffCommand(diffFilter, false), gitDir)
	if err != nil {
		return nil, err
	}

	return lines, nil
}
func parseGitZOutput(o string) []string {
	o = strings.TrimSuffix(o, "\x00")

	return strings.Split(o, "\x00")
}

func execGitZ(args []string, dir string) ([]string, error) {
	lines, err := execGit(args, dir)
	if err != nil {
		return nil, err
	}
	if lines == "" {
		return nil, nil
	}

	return parseGitZOutput(lines), nil
}

func getCachedFiles(gitDir string) ([]string, error) {
	return execGitZ([]string{"ls-files", "-z", "--full-name"}, gitDir)
}

func getTrackedFiles(diffFilter string, gitDir string) ([]string, error) {
	staged, err := getStagedFiles(diffFilter, gitDir)
	if err != nil {
		return nil, err
	}

	unstaged, err := getUnstagedFiles(diffFilter, gitDir)
	if err != nil {
		return nil, err
	}

	return uniqueStrings(append(staged, unstaged...)), nil
}

func getChangedFiles(diffFilter string, gitDir string) ([]string, error) {
	tracked, err := getTrackedFiles(diffFilter, gitDir)
	if err != nil {
		return nil, err
	}
	untracked, err := getUntrackedFiles(gitDir)
	if err != nil {
		return nil, err
	}

	return uniqueStrings(append(tracked, untracked...)), nil
}

func getUntrackedFiles(gitDir string) ([]string, error) {
	return execGitZ([]string{"ls-files", "-z", "--full-name", "--others", "--exclude-standard"}, gitDir)
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, item := range in {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}
