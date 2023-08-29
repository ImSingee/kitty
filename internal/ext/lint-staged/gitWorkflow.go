package lintstaged

import (
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/kitty/internal/lib/git"
	"log/slog"
	"path/filepath"
)

type gitWorkflow struct {
	root         string
	gitConfigDir string
	logger       *slog.Logger

	partiallyStagedFiles []string
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
func (g *gitWorkflow) getHiddenFilepath(filename string) string {
	return normalizePath(filepath.Join(g.gitConfigDir, filename))
}

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
func (g *gitWorkflow) prepare(state *State) (err error) {
	g.logger.Debug("Backing up original state...")

	g.partiallyStagedFiles, err = g.getPartiallyStagedFiles()
	if err != nil {
		return ee.Wrap(err, "cannot get partially staged files")
	}

	state.hasPartiallyStagedFiles = len(g.partiallyStagedFiles) > 0
	if state.hasPartiallyStagedFiles {
		unstagedPatch := g.getHiddenFilepath(PatchUnstaged)

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
		state.deletedFiles, err = g.getDeletedFiles()
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
		_, err = g.execGit("stash", "store", "--quiet", "--message", "lint-staged automatic backup", hash)
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

func (g *gitWorkflow) backupMergeStatus() error {
	g.logger.Debug("Backing up merge status...")
	g.logger.Warn("TODO: implement backupMergeStatus")

	// this.mergeHeadFilename = path.resolve(gitConfigDir, MERGE_HEAD)
	// this.mergeModeFilename = path.resolve(gitConfigDir, MERGE_MODE)
	// this.mergeMsgFilename = path.resolve(gitConfigDir, MERGE_MSG)

	// const MERGE_HEAD = 'MERGE_HEAD'
	// const MERGE_MODE = 'MERGE_MODE'
	// const MERGE_MSG = 'MERGE_MSG'

	// await Promise.all([
	//      readFile(this.mergeHeadFilename).then((buffer) => (this.mergeHeadBuffer = buffer)),
	//      readFile(this.mergeModeFilename).then((buffer) => (this.mergeModeBuffer = buffer)),
	//      readFile(this.mergeMsgFilename).then((buffer) => (this.mergeMsgBuffer = buffer)),
	//    ])

	g.logger.Debug("Done backing up merge state!")
	return nil
}
