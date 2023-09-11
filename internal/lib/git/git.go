package git

import (
	"errors"
	"fmt"
	"os/exec"
)

type G struct {
	Dir string
	Env []string
}

type Result struct {
	Output []byte

	ExitCode   int
	ExitErr    *exec.ExitError
	UnknownErr error
}

func (r *Result) Err() error {
	if r.ExitErr != nil {
		return fmt.Errorf("%s %s", r.ExitErr.Error(), r.ExitErr.Stderr)
	}

	if r.UnknownErr != nil {
		return r.UnknownErr
	}

	return nil
}

func (g *G) Run(args ...string) *Result {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	cmd.Env = g.Env
	output, err := cmd.Output()

	result := &Result{
		Output: output,
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			result.ExitErr = exitErr
		} else {
			result.ExitCode = -1
			result.UnknownErr = err
		}
	}

	return result
}

func run(dir string, args ...string) *Result {
	return (&G{Dir: dir}).Run(args...)
}

func R(wd string, args []string) *Result {
	return run(wd, args...)
}

func Run(args ...string) *Result {
	return run("", args...)
}
