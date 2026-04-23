package lintstaged

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSelectedFilesExcludesDeletedFilesForStatuses(t *testing.T) {
	repo := t.TempDir()

	gitRun(t, repo, "init")
	gitRun(t, repo, "config", "user.name", "Test User")
	gitRun(t, repo, "config", "user.email", "test@example.com")

	writeFile(t, repo, "clean.txt", "clean\n")
	writeFile(t, repo, "modified.txt", "before\n")
	writeFile(t, repo, "deleted-unstaged.txt", "gone\n")
	writeFile(t, repo, "deleted-staged.txt", "gone\n")

	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-m", "initial")

	writeFile(t, repo, "modified.txt", "after\n")
	writeFile(t, repo, "staged.txt", "staged\n")
	writeFile(t, repo, "untracked.txt", "untracked\n")
	require.NoError(t, os.Remove(filepath.Join(repo, "deleted-unstaged.txt")))
	gitRun(t, repo, "rm", "deleted-staged.txt")
	gitRun(t, repo, "add", "staged.txt")

	testCases := []struct {
		name     string
		options  *Options
		expected []string
	}{
		{
			name: "staged",
			options: &Options{
				Status:     string(SelectionModeStaged),
				DiffFilter: "ACMRD",
			},
			expected: []string{"staged.txt"},
		},
		{
			name: "unstaged",
			options: &Options{
				Status:     string(SelectionModeUnstaged),
				DiffFilter: "ACMRD",
			},
			expected: []string{"modified.txt"},
		},
		{
			name: "tracked",
			options: &Options{
				Status:     string(SelectionModeTracked),
				DiffFilter: "ACMRD",
			},
			expected: []string{"modified.txt", "staged.txt"},
		},
		{
			name: "changed",
			options: &Options{
				Status:     string(SelectionModeChanged),
				DiffFilter: "ACMRD",
			},
			expected: []string{"modified.txt", "staged.txt", "untracked.txt"},
		},
		{
			name: "all",
			options: &Options{
				Status: string(SelectionModeAll),
			},
			expected: []string{"clean.txt", "modified.txt", "staged.txt", "untracked.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := getSelectedFiles(tc.options, repo)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, files)
			assert.NotContains(t, files, "deleted-unstaged.txt")
			assert.NotContains(t, files, "deleted-staged.txt")
		})
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func writeFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
}
