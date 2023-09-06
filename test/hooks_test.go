package test

import (
	"bytes"
	"github.com/ImSingee/kitty/internal/lib/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestHooks(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		setup(t)

		kittyInstall(t)

		// check core.hooksPath
		expectHooksPathToBe(t, ".kitty")

		// test pre-commit
		runBash(t, "touch testfile && git add testfile")
		runBash(t, `kitty add pre-commit "echo \"pre-commit\" && exit 1"`)
		expectFailRunBash(t, "git commit -m foo")

		// uninstall
		runBash(t, "kitty uninstall")
		expectFailRunBash(t, "git config core.hooksPath")
		expectSuccessRunBash(t, "git commit -m foo")
	})

	t.Run("custom dir", func(t *testing.T) {
		setup(t)

		t.Skip("Not Supported Yet (TODO)")
	})

	t.Run("not git repo", func(t *testing.T) {
		setup(t)

		runBash(t, "rm -rf .git")
		expectFailRunBash(t, "kitty install")
	})

	t.Run("not git root", func(t *testing.T) {
		setup(t)

		runBash(t, "mkdir foo")
		expectFailRunBash(t, "cd foo && kitty install")
	})

	t.Run("git command not found", func(t *testing.T) {
		setup(t)

		expectSuccessRunBash(t, `
export KITTY=$(which kitty)

PATH='' $KITTY install
`)
	})

	t.Run("set and add hooks", func(t *testing.T) {
		setup(t)

		kittyInstall(t)

		t.Setenv("f", ".kitty/pre-commit")

		runBash(t, "kitty add pre-commit 'foo'")
		expectSuccessRunBash(t, "grep -m 1 _ $f && grep foo $f")

		runBash(t, "kitty add pre-commit 'bar'")
		expectSuccessRunBash(t, "grep -m 1 _ $f && grep foo $f && grep bar $f")

		runBash(t, "kitty set pre-commit 'baz'")
		expectSuccessRunBash(t, "grep -m 1 _ $f && grep baz $f")
		expectSuccessRunBash(t, `if grep -q foo "$f"; then exit 1; fi`)
		expectSuccessRunBash(t, `if grep -q bar "$f"; then exit 1; fi`)
	})

}

var buildKittyOnce sync.Once

func setup(t *testing.T) {
	t.Helper()

	buildKittyOnce.Do(func() {
		buildKitty(t)
	})

	testDir := "/tmp/kitty-test/" + t.Name()

	// create test directory
	err := os.RemoveAll(testDir)
	require.NoError(t, err)
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// enter test directory (and exit after test)
	previousWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(testDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Chdir(previousWd)
		require.NoError(t, err)
	})

	// init git for test directory
	gitRun(t, "init", "--quiet")
	gitRun(t, "config", "user.email", "test@test")
	gitRun(t, "config", "user.name", "test")
}

func getGitRoot(t *testing.T) string {
	t.Helper()

	return gitRun(t, "rev-parse", "--show-toplevel")
}

func expectSuccessRunBash(t *testing.T, script string) {
	t.Helper()

	runCommand(t, "bash", "-c", script)
}

func expectFailRunBash(t *testing.T, script string) {
	t.Helper()

	expectFailRunCommand(t, "bash", "-c", script)
}

func expectFailRunCommand(t *testing.T, args ...string) {
	t.Helper()

	cmd := exec.Command(args[0], args[1:]...)

	err := cmd.Run()
	require.Error(t, err)
	var exitError = &exec.ExitError{}
	require.ErrorAs(t, err, &exitError)
}

func runBash(t *testing.T, script string) string {
	t.Helper()

	return runCommand(t, "bash", "-c", script)
}

func runCommand(t *testing.T, args ...string) string {
	t.Helper()

	stdout := new(bytes.Buffer)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = stdout

	err := cmd.Run()
	require.NoError(t, err)

	return strings.TrimSpace(stdout.String())
}

func gitRun(t *testing.T, args ...string) string {
	t.Helper()

	result := git.Run(args...)
	require.NoError(t, result.Err())

	return strings.TrimSpace(string(result.Output))
}

func buildKitty(t *testing.T) {
	t.Helper()

	previousWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err = os.Chdir(previousWd)
		require.NoError(t, err)
	}()

	wd := getGitRoot(t)
	err = os.Chdir(wd)
	require.NoError(t, err)

	runCommand(t, "go", "build", "-o", ".kitty/.bin/kitty-test/kitty", "./cmd/kitty")

	err = os.Setenv("PATH", filepath.Join(wd, ".kitty/.bin/kitty-test")+":"+os.Getenv("PATH"))
	require.NoError(t, err)
}

func kittyInstall(t *testing.T) {
	t.Helper()

	runCommand(t, "kitty", "install")
}

func expectHooksPathToBe(t *testing.T, be string) {
	t.Helper()

	hooksPath := gitRun(t, "config", "core.hooksPath")
	assert.Equal(t, be, hooksPath)
}
