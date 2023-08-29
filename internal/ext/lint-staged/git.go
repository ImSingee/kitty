package lintstaged

import (
	"github.com/ImSingee/kitty/internal/lib/git"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	if diffFilter == "" {
		// Docs for --diff-filter option:
		// https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---diff-filterACDMRTUXB82308203
		diffFilter = "ACMR"
	}

	// Use `--diff branch1...branch2` or `--diff="branch1 branch2", or fall back to default staged files
	var diffArgs []string
	if diff == "" {
		diffArgs = []string{"--staged"}
	} else {
		diffArgs = strings.Split(strings.TrimSpace(diff), " ")
	}

	// Docs for -z option:
	// https://git-scm.com/docs/git-diff#Documentation/git-diff.txt--z
	return append([]string{"diff", "--name-only", "-z", "--diff-filter=" + diffFilter}, diffArgs...)
}

// getStagedFiles returns a list of staged files in relative path to git root
func getStagedFiles(options *Options, gitDir string) ([]string, error) {
	lines, err := execGitZ(getDiffCommand(options.Diff, options.DiffFilter), gitDir)
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
