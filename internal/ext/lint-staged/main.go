package lintstaged

import (
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"os"
)

type Options struct {
	AllowEmpty bool
	ConfigPath string
	Diff       string
	DiffFilter string
	Stash      bool
	Quiet      bool
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
	if options.Shell != "" {
		// TODO check is executable and have permission to execute
	} else {
		options.Shell = os.Getenv("SHELL")
		if options.Shell == "" {
			options.Shell = "/bin/sh"
		}
	}

	return nil
}
