package lintstaged

import (
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/kitty/internal/lib/tl"
	"log/slog"
	"os"
	"path/filepath"
)

func runAll(options *Options) (*State, error) {
	slog.Debug("Running all linter scripts...")

	cwd := options.Cwd
	err := error(nil)
	if cwd == "" {
		cwd, err = os.Getwd()
	} else {
		cwd, err = filepath.Abs(cwd)
	}
	if err != nil {
		return nil, ee.Wrap(err, "cannot get current working directory")
	}
	options.Cwd = cwd

	ctx := getInitialState(options)

	gitDir, gitConfigDir, err := resolveGitRepo(cwd)
	if err != nil {
		return ctx, ee.Wrap(err, "cannot resolve git repository")
	}
	if gitDir == "" {
		return ctx, ee.New("not a git repository")
	}

	// Test whether we have any commits or not.
	// Stashing must be disabled with no initial commit.
	_, err = execGit([]string{"log", "-1"}, gitDir)
	hasInitialCommit := err == nil

	// Lint-staged will create a backup stash only when there's an initial commit,
	// and when using the default list of staged files by default
	ctx.shouldBackup = hasInitialCommit && options.Stash
	if !ctx.shouldBackup {
		slog.Warn(skippingBackup(hasInitialCommit, options.Diff)) // TODO use print
	}

	files, err := getStagedFiles(options)
	if err != nil {
		return ctx, ee.Wrap(err, "cannot get staged files")
	}
	slog.Debug("Loaded list of staged files in git", "files", files)

	// If there are no files avoid executing any lint-staged logic
	if len(files) == 0 {
		if !options.Quiet {
			ctx.output = append(ctx.output, NO_STAGED_FILES)
		}
		return ctx, nil
	}

	foundConfigs, err := searchConfigs(cwd, gitDir, options.ConfigPath, nil) // TODO config object support
	if err != nil {
		return ctx, ee.Wrap(err, "cannot load configs")
	}

	if len(foundConfigs) == 0 {
		return ctx, ee.New("no configuration found") // TODO ErrCode ConfigNotFoundError
	}

	filesByConfig := groupFilesByConfig(foundConfigs, files)
	if debug() {
		usedConfigsCount := len(filesByConfig)
		debugFilesByConfig := make(map[string][]string, len(filesByConfig))

		for c, cFiles := range filesByConfig {
			debugFilesByConfig[c.Path] = cFiles
		}

		slog.Debug("Grouped staged files by config", "count", usedConfigsCount, "filesByConfig", debugFilesByConfig)
	}

	gw := newGitWorkflow(gitDir, gitConfigDir, slog.Default())

	tasks := []*tl.Task{
		{
			Title: "Preparing lint-staged...",
			Run: func(callback tl.TaskCallback) error {
				return gw.prepare(ctx)
			},
		},
		{
			Title: "Hiding unstaged changes to partially staged files...",
			Run: func(callback tl.TaskCallback) error {
				return nil // TODO git.hideUnstagedChanges(ctx),
			},
			Enable: func() bool {
				return ctx.hasPartiallyStagedFiles
			},
		},
		{
			Title: "Running tasks for staged files...",
			Run: func(callback tl.TaskCallback) error {
				callback.AddSubTaskList(tl.NewTaskList(
					[]*tl.Task{}, // TODO
					tl.WithExitOnError(true),
				))

				return nil // TODO concurrent
			},
			Options: []tl.OptionApplier{
				tl.WithExitOnError(false),
			},
		},
		{
			Title: "Applying modifications from tasks...",
			Run: func(callback tl.TaskCallback) error {
				//  skip: applyModificationsSkipped,

				return nil // TODO git.applyModifications(ctx),
			},
		},
		{
			Title: "Restoring unstaged changes to partially staged files...",
			Run: func(callback tl.TaskCallback) error {
				//  skip: restoreUnstagedChangesSkipped,

				return nil // TODO git.restoreUnstagedChanges(ctx),
			},
			Enable: func() bool {
				return ctx.hasPartiallyStagedFiles
			},
		},
		{
			Title: "Reverting to original state because of errors...",
			Run: func(callback tl.TaskCallback) error {
				//  skip: restoreOriginalStateSkipped

				return nil
			},
			Enable: func() bool {
				return false // restoreOriginalStateEnabled
			},
		},
		{
			Title: "Cleaning up temporary files...",
			Run: func(callback tl.TaskCallback) error {
				//  skip: cleanupSkipped

				return nil // TODO git.cleanup(ctx),
			},
			Enable: func() bool {
				return false // cleanupEnabled
			},
		},
	}

	tasks = mr.Filter(tasks, func(task *tl.Task, _ int) bool { return task != nil })
	runner := tl.New(tasks, tl.WithExitOnError(true))

	/*
	 [
	      {
	        title: 'Preparing lint-staged...',
	        task: (ctx) => git.prepare(ctx),
	      },
	      {
	        title: 'Hiding unstaged changes to partially staged files...',
	        task: (ctx) => git.hideUnstagedChanges(ctx),
	        enabled: hasPartiallyStagedFiles,
	      },
	      {
	        title: `Running tasks for staged files...`,
	        task: (ctx, task) => task.newListr(listrTasks, { concurrent }),
	        skip: () => listrTasks.every((task) => task.skip()),
	      },
	      {
	        title: 'Applying modifications from tasks...',
	        task: (ctx) => git.applyModifications(ctx),
	        skip: applyModificationsSkipped,
	      },
	      {
	        title: 'Restoring unstaged changes to partially staged files...',
	        task: (ctx) => git.restoreUnstagedChanges(ctx),
	        enabled: hasPartiallyStagedFiles,
	        skip: restoreUnstagedChangesSkipped,
	      },
	      {
	        title: 'Reverting to original state because of errors...',
	        task: (ctx) => git.restoreOriginalState(ctx),
	        enabled: restoreOriginalStateEnabled,
	        skip: restoreOriginalStateSkipped,
	      },
	      {
	        title: 'Cleaning up temporary files...',
	        task: (ctx) => git.cleanup(ctx),
	        enabled: cleanupEnabled,
	        skip: cleanupSkipped,
	      },
	    ],
	*/

	/*

	  const hasMultipleConfigs = numberOfConfigs > 1

	  // lint-staged 10 will automatically add modifications to index
	  // Warn user when their command includes `git add`
	  let hasDeprecatedGitAdd = false

	  const listrTasks = []

	  // Set of all staged files that matched a task glob. Values in a set are unique.
	  const matchedFiles = new Set()

	  for (const [configPath, { config, files }] of Object.entries(filesByConfig)) {
	    const configName = configPath ? normalizePath(path.relative(cwd, configPath)) : 'Config object'

	    const stagedFileChunks = chunkFiles({ baseDir: gitDir, files, maxArgLength, relative })

	    // Use actual cwd if it's specified, or there's only a single config file.
	    // Otherwise use the directory of the config file for each config group,
	    // to make sure tasks are separated from each other.
	    const groupCwd = hasMultipleConfigs && !hasExplicitCwd ? path.dirname(configPath) : cwd

	    const chunkCount = stagedFileChunks.length
	    if (chunkCount > 1) {
	      debugLog('Chunked staged files from `%s` into %d part', configPath, chunkCount)
	    }

	    for (const [index, files] of stagedFileChunks.entries()) {
	      const chunkListrTasks = await Promise.all(
	        generateTasks({ config, cwd: groupCwd, files, relative }).map((task) =>
	          makeCmdTasks({
	            commands: task.commands,
	            cwd: groupCwd,
	            files: task.fileList,
	            gitDir,
	            shell,
	            verbose,
	          }).then((subTasks) => {
	            // Add files from task to match set
	            task.fileList.forEach((file) => {
	              // Make sure relative files are normalized to the
	              // group cwd, because other there might be identical
	              // relative filenames in the entire set.
	              const normalizedFile = path.isAbsolute(file)
	                ? file
	                : normalizePath(path.join(groupCwd, file))

	              matchedFiles.add(normalizedFile)
	            })

	            hasDeprecatedGitAdd =
	              hasDeprecatedGitAdd || subTasks.some((subTask) => subTask.command === 'git add')

	            const fileCount = task.fileList.length

	            return {
	              title: `${task.pattern}${chalk.dim(
	                ` — ${fileCount} ${fileCount === 1 ? 'file' : 'files'}`
	              )}`,
	              task: async (ctx, task) =>
	                task.newListr(
	                  subTasks,
	                  // Subtasks should not run in parallel, and should exit on error
	                  { concurrent: false, exitOnError: true }
	                ),
	              skip: () => {
	                // Skip task when no files matched
	                if (fileCount === 0) {
	                  return `${task.pattern}${chalk.dim(' — no files')}`
	                }
	                return false
	              },
	            }
	          })
	        )
	      )

	      listrTasks.push({
	        title:
	          `${configName}${chalk.dim(` — ${files.length} ${files.length > 1 ? 'files' : 'file'}`)}` +
	          (chunkCount > 1 ? chalk.dim(` (chunk ${index + 1}/${chunkCount})...`) : ''),
	        task: (ctx, task) => task.newListr(chunkListrTasks, { concurrent, exitOnError: true }),
	        skip: () => {
	          // Skip if the first step (backup) failed
	          if (ctx.errors.has(GitError)) return SKIPPED_GIT_ERROR
	          // Skip chunk when no every task is skipped (due to no matches)
	          if (chunkListrTasks.every((task) => task.skip())) {
	            return `${configName}${chalk.dim(' — no tasks to run')}`
	          }
	          return false
	        },
	      })
	    }
	  }



	*/

	err = runner.Run()

	return ctx, err
}
