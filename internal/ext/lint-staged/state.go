package lintstaged

import (
	"github.com/ImSingee/go-ex/set"
	"sync"
)

type State struct {
	shouldBackup bool
	quiet        bool

	hasPartiallyStagedFiles bool
	taskResults             *sync.Map

	output        []string
	errors        *set.Set[error]
	internalError bool
	taskError     bool
}

type TaskResult struct {
	cmd                *Command
	fullCommandAndArgs string
	output             []byte
	err                error
}

func getInitialState(options *Options) *State {
	return &State{
		hasPartiallyStagedFiles: false,
		taskResults:             &sync.Map{},
		errors:                  set.New[error](),
		quiet:                   options.Quiet,
	}
}
