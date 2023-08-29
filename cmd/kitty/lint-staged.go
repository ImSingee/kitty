package main

import (
	lintstaged "github.com/ImSingee/kitty/internal/ext/lint-staged"
	"github.com/spf13/cobra"
)

func init() {
	var allowEmpty bool
	var configPath string
	var cwd string
	var diff string
	var diffFilter string
	var stash bool
	var quiet bool
	var relative bool
	var shell string
	var verbose bool

	cmd := &cobra.Command{
		Use:     "lint-staged",
		Aliases: []string{"lint"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if diff != "" {
				stash = false
			}
			if configPath == "-" { // read from stdin
				// TODO
			}

			options := &lintstaged.Options{
				AllowEmpty: allowEmpty,
				ConfigPath: configPath,
				Cwd:        cwd,
				Debug:      debug,
				Diff:       diff,
				DiffFilter: diffFilter,
				Stash:      stash,
				Quiet:      quiet,
				Relative:   relative,
				Shell:      shell,
				Verbose:    verbose,
			}

			return lintstaged.Run(options)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&allowEmpty, "allow-empty", false, "allow empty commits when tasks revert all staged changes")
	flags.StringVarP(&configPath, "config", "c", "", "path to configuration file, or - to read from stdin")
	flags.StringVar(&cwd, "cwd", "", "run all tasks in specific directory, instead of the current")
	flags.StringVar(&diff, "diff", "", `override the default "--staged" flag of "git diff" to get list of files. Implies "--stash=false"`)
	flags.StringVar(&diffFilter, "diff-filter", "", `override the default "--diff-filter=ACMR" flag of "git diff" to get list of files`)
	flags.BoolVar(&stash, "stash", true, "enable the backup stash, and revert in case of errors")
	flags.BoolVarP(&quiet, "quiet", "q", false, "disable lint-staged’s own console output")
	flags.BoolVarP(&relative, "relative", "r", false, "pass relative filepaths to tasks")
	flags.StringVarP(&shell, "shell", "x", "", "skip parsing of tasks for better shell support")
	flags.BoolVarP(&verbose, "verbose", "v", false, "show task output even when tasks succeed; by default only failed output is shown")

	// TODO concurrent

	extensions = append(extensions, cmd)
}
