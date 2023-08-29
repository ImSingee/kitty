package lintstaged

import (
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/kitty/internal/lib/git"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type gitWorkflow struct {
	root         string
	gitConfigDir string
	logger       *slog.Logger

	partiallyStagedFiles []string
	deletedFiles         []string
	mergeStatusBackup    map[string][]byte
}

func newGitWorkflow(root string, gitConfigDir string, logger *slog.Logger) *gitWorkflow {
	return &gitWorkflow{
		root:         root,
		gitConfigDir: gitConfigDir,
		logger:       logger,
	}
}

func (g *gitWorkflow) execGit(args ...string) (string, error) {
	return execGit(args, g.root)
}

func (g *gitWorkflow) execGitZ(args ...string) ([]string, error) {
	return execGitZ(args, g.root)
}

func (g *gitWorkflow) execStatus() ([]git.FileStatus, error) {
	return git.Status(g.root, false)
}

// Get absolute path to file hidden inside .git
func (g *gitWorkflow) getGitConfigDirFilepath(filename string) string {
	return normalizePath(filepath.Join(g.gitConfigDir, filename))
}

const stashMessage = "lint-staged automatic backup"
const PatchUnstaged = "lint-staged_unstaged.patch"

var gitDiffArgs = []string{
	"--binary",          // support binary files
	"--unified=0",       // do not add lines around diff for consistent behaviour
	"--no-color",        // disable colors for consistent behaviour
	"--no-ext-diff",     // disable external diff tools for consistent behaviour
	"--src-prefix=a/",   // force prefix for consistent behaviour
	"--dst-prefix=b/",   // force prefix for consistent behaviour
	"--patch",           // output a patch that can be applied
	"--submodule=short", // always use the default short format for submodules
}
var gitApplyArgs = []string{"-v", "--whitespace=nowarn", "--recount", "--unidiff-zero"}

// Create a diff of partially staged files and backup stash if enabled.
//
// will set state.hasPartiallyStagedFiles
func (g *gitWorkflow) prepare(state *State) (err error) {
	g.logger.Debug("Backing up original state...")

	g.partiallyStagedFiles, err = g.getPartiallyStagedFiles()
	if err != nil {
		return ee.Wrap(err, "cannot get partially staged files")
	}

	state.hasPartiallyStagedFiles = len(g.partiallyStagedFiles) > 0
	if state.hasPartiallyStagedFiles {
		unstagedPatch := g.getGitConfigDirFilepath(PatchUnstaged)

		args := mr.Flats(
			[]string{"diff"},
			gitDiffArgs,
			[]string{"--output", unstagedPatch, "--"},
			g.partiallyStagedFiles,
		)

		_, err := g.execGit(args...)
		if err != nil {
			return ee.Wrap(err, "cannot create patch for partially staged files")
		}
	}

	if state.shouldBackup {
		// When backup is enabled, the revert will clear ongoing merge status.
		err := g.backupMergeStatus()
		if err != nil {
			return ee.Wrap(err, "cannot backup merge status")
		}

		// Get a list of unstaged deleted files, because certain bugs might cause them to reappear:
		// - in git versions =< 2.13.0 the `git stash --keep-index` option resurrects deleted files
		// - git stash can't infer RD or MD states correctly, and will lose the deletion
		g.deletedFiles, err = g.getDeletedFiles()
		if err != nil {
			return ee.Wrap(err, "cannot get deleted files")
		}

		// Save stash of all staged files.
		// The `stash create` command creates a dangling commit without removing any files,
		// and `stash store` saves it as an actual stash.
		hash, err := g.execGit("stash", "create")
		if err != nil {
			return ee.Wrap(err, "cannot create stash")
		}
		_, err = g.execGit("stash", "store", "--quiet", "--message", stashMessage, hash)
		if err != nil {
			return ee.Wrap(err, "cannot save stash")
		}

		g.logger.Debug("Done backing up original state!")
	}

	return nil
}

// Get a list of all files with both staged and unstaged modifications.
//
// Renames have special treatment, since the single status line includes
// both the "from" and "to" filenames, where "from" is no longer on disk.
// So, for rename, only the "to" is returned.
func (g *gitWorkflow) getPartiallyStagedFiles() ([]string, error) {
	g.logger.Debug("Getting partially staged files...")

	status, err := g.execStatus()
	if err != nil {
		return nil, ee.Wrap(err, "get git status failed")
	}

	status = mr.Filter(status, func(status git.FileStatus, _ int) bool {
		return status.IndexStatus != ' ' && status.WorkingTreeStatus != ' ' && status.IndexStatus != '?' && status.WorkingTreeStatus != '?'
	})
	files := mr.Map(status, func(status git.FileStatus, _ int) string {
		return status.Name
	})

	g.logger.Debug("Found partially staged files", "files", files)

	return files, nil
}

// Get a list of unstaged deleted files
func (g *gitWorkflow) getDeletedFiles() ([]string, error) {
	g.logger.Debug("Getting deleted files...")

	files, err := g.execGitZ("ls-files", "--deleted", "-z")
	if err != nil {
		return nil, err
	}

	g.logger.Debug("Found deleted files", "files", files)

	// convert to absolute
	files = mr.Map(files, func(file string, _ int) string {
		return filepath.Join(g.root, file)
	})

	return files, nil
}

// backupMergeStatus backup merge info to state
func (g *gitWorkflow) backupMergeStatus() (err error) {
	g.logger.Debug("Backing up merge status...")

	g.mergeStatusBackup, err = readFiles(
		g.getGitConfigDirFilepath("MERGE_HEAD"),
		g.getGitConfigDirFilepath("MERGE_MODE"),
		g.getGitConfigDirFilepath("MERGE_MSG"),
	)
	if err != nil {
		return ee.Wrap(err, "cannot backup merge status")
	}

	g.logger.Debug("Done backing up merge state!")
	return nil
}

// will ignore not-exist file
func readFiles(filenames ...string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(filenames))
	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err == nil {
			result[filename] = data
		} else if os.IsNotExist(err) {
			continue
		} else {
			return nil, err
		}
	}

	return result, nil
}

func writeFiles(record map[string][]byte) error {
	for filename, data := range record {
		err := os.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// Restore meta information about ongoing git merge
//
// will add ErrRestoreMergeStatus to state.errors if fail
func (g *gitWorkflow) restoreMergeStatus(state *State) (err error) {
	defer func() {
		if err != nil {
			state.errors.Add(ErrRestoreMergeStatus)
		}
	}()

	g.logger.Debug("Restoring merge state...")

	err = writeFiles(g.mergeStatusBackup)
	if err != nil {
		return ee.Wrap(err, "cannot restore merge status")
	}

	g.logger.Debug("Done backing up merge state!")
	return nil
}

// Remove unstaged changes to all partially staged files, to avoid tasks from seeing them
func (g *gitWorkflow) hideUnstagedChanges() error {
	_, err := g.execGit(append([]string{"checkout", "--force", "--"}, g.partiallyStagedFiles...)...)
	if err != nil {
		// `git checkout --force` doesn't throw errors, so it shouldn't be possible to get here.
		return err
	}
	return nil
}

// Restore unstaged changes to partially changed files.
// If it at first fails, this is probably because of conflicts between new task modifications.
// 3-way merge usually fixes this, and in case it doesn't we should just give up and throw.
func (g *gitWorkflow) restoreUnstagedChanges() error {
	g.logger.Debug("Restoring unstaged changes...")

	unstagedPatch := g.getGitConfigDirFilepath(PatchUnstaged)

	_, applyErr := g.execGit(mr.Flats([]string{"apply"}, gitApplyArgs, []string{unstagedPatch})...)
	if applyErr == nil {
		return nil
	}

	g.logger.Debug("Error while restoring changes", "error", applyErr)

	g.logger.Debug("Retrying with 3-way merge...")

	_, threeWayApplyErr := g.execGit(mr.Flats([]string{"apply"}, gitApplyArgs, []string{"--3way", unstagedPatch})...)
	if threeWayApplyErr == nil {
		return nil
	}

	g.logger.Debug("Error while restoring changes using 3-way merge", "error", threeWayApplyErr)

	return ee.New("unstaged changes could not be restored due to a merge conflict")
}

// Restore original HEAD state in case of errors
func (g *gitWorkflow) restoreOriginalState(state *State) (err error) {
	g.logger.Debug("Restoring original state...")

	backupStash, err := g.getBackupStashIndex()
	if err != nil {
		return ee.Wrap(err, "cannot get backup stash index")
	}

	_, err = g.execGit("reset", "--hard", "HEAD")
	if err != nil {
		return err
	}
	_, err = g.execGit("stash", "apply", "--quiet", "--index", backupStash)
	if err != nil {
		return err
	}

	err = g.restoreMergeStatus(state)
	if err != nil {
		return ee.Wrap(err, "cannot restore merge status")
	}

	// If stashing resurrected deleted files, clean them out
	removeFilesAndIgnoreError(g.deletedFiles)

	// Clean out patch
	_ = os.Remove(g.getGitConfigDirFilepath(PatchUnstaged))

	g.logger.Debug("Done restoring original state!")
	return nil
}

func removeFilesAndIgnoreError(files []string) {
	for _, file := range files {
		_ = os.Remove(file)
	}
}

// Get name of backup stash
func (g *gitWorkflow) getBackupStashIndex() (string, error) {
	stashes, err := g.execGitZ("stash", "list", "-z")
	if err != nil {
		return "", ee.Wrap(err, "cannot get stash list")
	}

	index := mr.FindIndex(stashes, func(stash string) bool {
		return strings.Contains(stash, stashMessage)
	})

	if index == -1 {
		return "", ee.New("miss lint-staged automatic backup")
	}

	return strconv.Itoa(index), nil
}

// Drop the created stashes after everything has run
func (g *gitWorkflow) cleanup() error {
	g.logger.Debug("Dropping backup stash...")

	stash, err := g.getBackupStashIndex()
	if err != nil {
		return ee.Wrap(err, "cannot get backup stash index")
	}

	_, err = g.execGit("stash", "drop", "--quiet", stash)
	if err != nil {
		return ee.Wrap(err, "cannot drop backup stash")
	}

	g.logger.Debug("Done dropping backup stash!")
	return nil
}
