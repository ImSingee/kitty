package git

import (
	"errors"
	"os/exec"
)

type Result struct {
	Output []byte

	ExitCode   int
	ExitErr    *exec.ExitError
	UnknownErr error
}

func (r *Result) Err() error {
	if r.ExitErr != nil {
		return r.ExitErr
	}

	if r.UnknownErr != nil {
		return r.UnknownErr
	}

	return nil
}

func run(args ...string) *Result {
	output, err := exec.Command("git", args...).Output()

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

func Run(args ...string) *Result {
	return run(args...)
}
