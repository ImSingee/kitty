package lintstaged

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "@lint-staged",
		Aliases: []string{"@lintstaged"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.Diff != "" {
				o.Stash = false
			}

			return Run(o)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVar(&o.AllowEmpty, "allow-empty", false, "allow empty commits when tasks revert all staged changes")
	flags.StringVarP(&o.ConfigPath, "config", "c", "", "path to configuration file")
	flags.StringVar(&o.Diff, "diff", "", `override the default "--staged" flag of "git diff" to get list of files. Implies "--stash=false"`)
	flags.StringVar(&o.DiffFilter, "diff-filter", "", `override the default "--diff-filter=ACMR" flag of "git diff" to get list of files`)
	flags.BoolVar(&o.Stash, "stash", true, "enable the backup stash, and revert in case of errors")
	flags.StringVarP(&o.Shell, "shell", "x", "", "use a custom shell to execute tasks with; defaults to the shell specified in the environment variable $SHELL, or /bin/sh if not set")
	flags.BoolVarP(&o.Verbose, "verbose", "v", false, "show task output even when tasks succeed; by default only failed output is shown")

	// TODO concurrent

	return []*cobra.Command{cmd}
}

type Options struct {
	AllowEmpty bool
	ConfigPath string
	Diff       string
	DiffFilter string
	Stash      bool
	Shell      string
	Verbose    bool
}

func Run(options *Options) error {
	if err := validateOptions(options); err != nil {
		return ee.Wrap(err, "invalid options")
	}

	// Unset GIT_LITERAL_PATHSPECS to not mess with path interpretation
	unsetEnv("GIT_LITERAL_PATHSPECS")

	state, err := runAll(options)

	for _, output := range state.output {
		pp.EPrintln(output)
	}

	return err
}

func validateOptions(options *Options) error {
	if options.Shell == "" {
		options.Shell = os.Getenv("SHELL")
		if options.Shell == "" {
			options.Shell = "/bin/sh"
		}
	}
	shell, err := exec.LookPath(options.Shell)
	if err != nil {
		return fmt.Errorf("shell `%s` not found or cannot execute", options.Shell)
	}
	options.Shell = shell

	return nil
}
