package lintstaged

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exbytes"
	"github.com/ImSingee/go-ex/glob"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/pp"

	"github.com/ImSingee/kitty/internal/lib/shells"
	"github.com/ImSingee/kitty/internal/lib/tl"
)

func runAll(options *Options) (*State, error) {
	slog.Debug("Running all linter scripts...")

	cwd, err := os.Getwd()
	if err != nil {
		return nil, ee.Wrap(err, "cannot get current working directory")
	}

	ctx := getInitialState(options)

	gitDir, gitConfigDir, err := resolveGitRepo(cwd)
	if err != nil || gitDir == "" {
		ctx.errors.Add(ErrGitRepo)
		pp.ERedPrintln(x, "Current directory is not a git directory!")
		return ctx, ee.Phantom
	}

	// Test whether we have any commits or not.
	// Stashing must be disabled with no initial commit.
	_, err = execGit([]string{"log", "-1"}, gitDir)
	hasInitialCommit := err == nil

	// Lint-staged will create a backup stash only when there's an initial commit,
	// and when using the default list of staged files by default
	ctx.shouldBackup = hasInitialCommit && options.Stash
	if !ctx.shouldBackup {
		pp.EYellowPrintln(skippingBackup(hasInitialCommit, options.Diff))
	}

	// get staged files (relative path)
	relativeFiles, err := getStagedFiles(options, gitDir)
	if err != nil {
		ctx.errors.Add(ErrGetStagedFiles)
		return ctx, ee.Wrap(err, "cannot get staged files")
	}

	slog.Debug("Loaded list of staged files in git (relative)", "files", relativeFiles)

	absoluteFiles := mr.Map(relativeFiles, func(in string, _index int) string {
		return normalizePath(filepath.Join(gitDir, in))
	})

	slog.Debug("Loaded list of staged files in git (absolute)", "files", absoluteFiles)

	// If there are no files avoid executing any lint-staged logic
	if len(absoluteFiles) == 0 {
		pp.BluePrintln(info, "No staged files found.")
		return ctx, nil
	}

	foundConfigs, err := searchConfigs(cwd, gitDir, options.ConfigPath)
	if err != nil {
		return ctx, ee.Wrap(err, "cannot load configs")
	}

	if len(foundConfigs) == 0 {
		ctx.errors.Add(ErrConfigNotFound)
		return ctx, ee.New("no configuration found")
	}

	filesByConfig := groupFilesByConfig(foundConfigs, absoluteFiles)
	if debug() {
		usedConfigsCount := len(filesByConfig)
		debugFilesByConfig := make(map[string][]string, len(filesByConfig))

		for c, cFiles := range filesByConfig {
			debugFilesByConfig[c.Path] = cFiles
		}

		slog.Debug("Grouped staged files by config", "count", usedConfigsCount, "filesByConfig", debugFilesByConfig)
	}

	chunkedFilenamesArray := chunkFiles(relativeFiles, defaultMaxArgLength())
	slog.Debug("Get chunked filenames arrays", "groupCount", len(chunkedFilenamesArray), "arrays", chunkedFilenamesArray)

	gw := &gitWorkflow{
		root:                  gitDir,
		gitConfigDir:          gitConfigDir,
		chunkedFilenamesArray: chunkedFilenamesArray,
		allowEmpty:            options.AllowEmpty,
		diff:                  options.Diff,
		diffFilter:            options.DiffFilter,
		logger:                slog.Default(), // TODO
	}

	handleInternalError := func(result *tl.Result) {
		if result.Error {
			ctx.internalError = true
		}
	}
	handleInternalErrorAnd := func(then func(result *tl.Result)) func(result *tl.Result) {
		return func(result *tl.Result) {
			if result.Error {
				ctx.internalError = true
				then(result)
			}
		}
	}
	handleTaskError := func(result *tl.Result) {
		if result.Error {
			ctx.taskError = true
		}
	}

	subTasks := generateTasksToRun(ctx, filesByConfig, options)

	tasks := []*tl.Task{
		{
			Title: "Preparing lint-staged...",
			Run: func(callback tl.TaskCallback) error {
				return gw.prepare(ctx)
			},
			PostRun: handleInternalError,
		},
		{
			Title: "Hiding unstaged changes to partially staged files...",
			Run: func(callback tl.TaskCallback) error {
				return gw.hideUnstagedChanges()
			},
			Enable: func() bool {
				return ctx.hasPartiallyStagedFiles
			},
			PostRun: handleInternalErrorAnd(func(result *tl.Result) {
				ctx.errors.Add(ErrHideUnstagedChanges)
			}),
		},
		{
			Title: "Running tasks for staged files...",
			Run: func(callback tl.TaskCallback) error {
				if ctx.internalError {
					callback.Skip("internal error")
					return nil
				}

				callback.AddSubTaskList(tl.NewTaskList(
					subTasks,
					tl.WithExitOnError(true),
				))

				return nil // TODO concurrent
			},
			PostRun: handleTaskError,
			Options: []tl.OptionApplier{
				tl.WithExitOnError(false),
			},
		},
		{
			Title: "Applying modifications from tasks...",
			Run: func(callback tl.TaskCallback) error {
				if ctx.shouldBackup { // Always apply back unstaged modifications when skipping backup
					if ctx.internalError {
						callback.Skip("internal error")
						return nil
					}
					if ctx.taskError {
						callback.Skip("task error")
						return nil
					}
				}

				return gw.applyModifications(ctx)
			},
			PostRun: handleInternalError,
		},
		{
			Title: "Restoring unstaged changes to partially staged files...",
			Run: func(callback tl.TaskCallback) error {
				// if error, skip and use "Reverting to original state because of errors" step
				if ctx.internalError {
					callback.Skip("internal error")
					return nil
				}
				if ctx.taskError {
					callback.Skip("task error")
					return nil
				}

				return gw.restoreUnstagedChanges()
			},
			Enable: func() bool {
				return ctx.hasPartiallyStagedFiles
			},
			PostRun: handleInternalErrorAnd(func(result *tl.Result) {
				ctx.errors.Add(ErrRestoreUnstagedChanges)
			}),
		},
		{
			Title: "Reverting to original state because of errors...",
			Run: func(callback tl.TaskCallback) error {
				if ctx.internalError {
					if ctx.errors.Has(ErrApplyEmptyCommit) || ctx.errors.Has(ErrRestoreUnstagedChanges) {
						// continue
					} else {
						callback.Skip("internal error")
						return nil
					}
				}

				return gw.restoreOriginalState(ctx)
			},
			Enable: func() bool {
				return ctx.shouldBackup && (ctx.taskError || ctx.errors.Has(ErrApplyEmptyCommit) || ctx.errors.Has(ErrRestoreUnstagedChanges))
			},
			PostRun: handleInternalErrorAnd(func(result *tl.Result) {
				ctx.errors.Add(ErrRestoreOriginalState)
			}),
		},
		{
			Title: "Cleaning up temporary files...",
			Run: func(callback tl.TaskCallback) error {
				// skip if previous (not - task) error
				if ctx.internalError {
					if ctx.errors.Has(ErrApplyEmptyCommit) || ctx.errors.Has(ErrRestoreUnstagedChanges) {
						// continue
					} else {
						callback.Skip("internal error")
						return nil
					}
				}

				return gw.cleanup()
			},
			Enable: func() bool {
				return ctx.shouldBackup
			},
		},
	}

	runner := tl.New(tasks, tl.WithExitOnError(true))
	err = runner.Run()

	printTaskResults(ctx.taskResults, options)

	return ctx, err
}

func generateTasksToRun(ctx *State, config map[*Config][]string, options *Options) []*tl.Task {
	type ConfigEntries struct {
		Config    *Config
		Filenames []string
		Tasks     []*tl.Task
	}

	all := make([]*ConfigEntries, 0, len(config))

	for config, files := range config {
		tasks := generateTasksForConfig(ctx, config, files, options)

		all = append(all, &ConfigEntries{
			Config:    config,
			Filenames: files,
			Tasks:     tasks,
		})
	}

	sort.Slice(all, func(i, j int) bool {
		return strings.Compare(all[i].Config.Path, all[j].Config.Path) < 0
	})

	if len(all) == 1 {
		return all[0].Tasks
	}

	return mr.Map(all, func(in *ConfigEntries, index int) *tl.Task {
		return &tl.Task{
			Title: in.Config.Path + symGray(fmt.Sprintf(" - %d files", len(in.Filenames))),
			Run: func(callback tl.TaskCallback) error {
				callback.AddSubTask(in.Tasks...)
				return nil
			},
		}
	})
}

func generateTasksForConfig(ctx *State, config *Config, files []string, options *Options) []*tl.Task {
	type RuleConfigEntries struct {
		Rule string
	}

	if len(config.Rules) == 0 {
		return []*tl.Task{{
			Title: "No Rules",
			Run: func(callback tl.TaskCallback) error {
				return nil
			},
		}}
	}

	wd := filepath.Dir(config.Path)

	if len(config.Rules) == 1 {
		return []*tl.Task{generateTaskForRule(ctx, wd, config.Rules[0], files, options)}
	}

	return mr.Map(config.Rules, func(rule *Rule, index int) *tl.Task {
		return generateTaskForRule(ctx, wd, rule, files, options)
	})
}

func generateTaskForRule(ctx *State, wd string, rule *Rule, files []string, options *Options) *tl.Task {
	files = mr.Filter(files, func(in string, index int) bool {
		return strings.HasPrefix(in, wd+string(filepath.Separator))
	})
	files = mr.Filter(files, func(in string, index int) bool {
		// TODO support multi level match
		return glob.Match(rule.Glob, filepath.Base(in))
	})

	suffix := fmt.Sprintf(" - %d files", len(files))
	if len(files) == 0 {
		suffix = " - no files"
	}

	cmdTasks := mr.Map(rule.Commands, func(cmd *Command, index int) *tl.Task {
		return generateTaskForCommand(ctx, wd, cmd, files, options)
	})

	return &tl.Task{
		Title: rule.Glob + symGray(suffix),
		Run: func(callback tl.TaskCallback) error {
			if len(files) == 0 {
				callback.Skip("")
				return nil
			}

			callback.AddSubTask(cmdTasks...)
			return nil
		},
		PostRun: func(result *tl.Result) {
			if !result.Error && !options.Verbose {
				result.Hide = true
			}
		},
	}
}

func generateTaskForCommand(state *State, wd string, cmd *Command, onFiles []string, options *Options) *tl.Task {
	return &tl.Task{
		Title: cmd.Command + symGray(fmt.Sprintf(" - %d files", len(onFiles))),
		Run: func(callback tl.TaskCallback) (err error) {
			if len(onFiles) == 0 {
				callback.Skip("")
				return nil
			}

			shell := options.Shell

			fileArgs := ""
			if cmd.NoArgs {
				// do nothing
			} else if cmd.Absolute {
				fileArgs = shells.Join(onFiles)
			} else {
				fileArgs = shells.Join(mr.Map(onFiles, func(f string, index int) string {
					return strings.TrimPrefix(f, wd+string(filepath.Separator))
				}))
			}

			args := cmd.execCommand + " " + fileArgs

			p := exec.Command(shell, "-c", args)
			p.Dir = wd

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			go func() {
				<-ctx.Done()

				if process := p.Process; process != nil {
					_ = process.Kill()
				}
			}()

			output, err := p.CombinedOutput()

			result := &TaskResult{
				cmd:                cmd,
				fullCommandAndArgs: args,
				output:             output,
				err:                err,
			}
			state.taskResults.Store(callback.GetTask().Id(), result)

			return err
		},
	}
}

func printTaskResults(results *sync.Map, options *Options) {
	anyResult := false
	results.Range(func(key, value any) bool {
		result := value.(*TaskResult)

		success := result.err == nil

		if success && !options.Verbose {
			return true
		}

		if !anyResult { // first print
			fmt.Println()
			anyResult = true
		}

		icon := x
		if success {
			icon = yes
		}

		cmd := result.cmd.Command

		output := exbytes.ToString(bytes.TrimSpace(result.output))
		output = strings.ToValidUTF8(output, "\uFFFD")

		if len(output) == 0 {
			if success {
				pp.GreenPrintf("%s %s success without output.\n", icon, cmd)
			} else {
				pp.RedPrintf("%s %s failed without output. (%v)\n", icon, cmd, result.err)
			}
		} else {
			if success {
				pp.GreenPrintf("%s %s:\n", icon, cmd)
			} else {
				pp.RedPrintf("%s %s:\n", icon, cmd)
			}

			pp.Println(output)
		}

		return true
	})

	if anyResult {
		fmt.Println()
	}
}
