package lintstaged

import (
	"fmt"
	"github.com/ImSingee/go-ex/ee"
)

type Options struct {
	AllowEmpty bool
	ConfigPath string
	Cwd        string
	Debug      bool
	Diff       string
	DiffFilter string
	Stash      bool
	Quiet      bool
	Relative   bool
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
		fmt.Println(output)
	}

	_ = state // TODO create output

	return err
}

func validateOptions(options *Options) error {
	if options.Cwd != "" {
		// TODO check is folder and have permission to access
	}

	if options.Shell != "" {
		// TODO check is executable and have permission to execute
	}

	return nil
}
