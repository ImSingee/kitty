package lintstaged

import (
	"sync"

	"github.com/ImSingee/go-ex/set"
)

type State struct {
	wd           string
	gitRoot      string
	shouldBackup bool

	hasPartiallyStagedFiles bool
	taskResults             *sync.Map
	ignoreChecker           *IgnoreChecker

	output        []string // all outputs will print to stderr at end
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

func getInitialState(wd string, options *Options) *State {
	return &State{
		wd:                      wd,
		hasPartiallyStagedFiles: false,
		taskResults:             &sync.Map{},
		errors:                  set.New[error](),
	}
}
