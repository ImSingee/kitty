package main

import (
	lintstaged "github.com/ImSingee/kitty/internal/ext/lint-staged"
	"github.com/spf13/cobra"
)

func init() {
	var allowEmpty bool
	var configPath string
	var diff string
	var diffFilter string
	var stash bool
	var shell string
	var verbose bool

	cmd := &cobra.Command{
		Use:     "lint-staged",
		Aliases: []string{"lint", "@lint-staged"},
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
				Diff:       diff,
				DiffFilter: diffFilter,
				Stash:      stash,
				Shell:      shell,
				Verbose:    verbose,
			}

			return lintstaged.Run(options)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&allowEmpty, "allow-empty", false, "allow empty commits when tasks revert all staged changes")
	flags.StringVarP(&configPath, "config", "c", "", "path to configuration file, or - to read from stdin")
	flags.StringVar(&diff, "diff", "", `override the default "--staged" flag of "git diff" to get list of files. Implies "--stash=false"`)
	flags.StringVar(&diffFilter, "diff-filter", "", `override the default "--diff-filter=ACMR" flag of "git diff" to get list of files`)
	flags.BoolVar(&stash, "stash", true, "enable the backup stash, and revert in case of errors")
	flags.StringVarP(&shell, "shell", "x", "", "use a custom shell to execute tasks with; defaults to the shell specified in the environment variable $SHELL, or /bin/sh if not set")
	flags.BoolVarP(&verbose, "verbose", "v", false, "show task output even when tasks succeed; by default only failed output is shown")

	// TODO concurrent

	extensions = append(extensions, cmd)
}
