package lintstaged

import "github.com/ImSingee/go-ex/set"

type State struct {
	shouldBackup bool
	quiet        bool

	hasPartiallyStagedFiles bool

	output        []string
	errors        *set.Set[error]
	internalError bool
	taskError     bool
}

func getInitialState(options *Options) *State {
	return &State{
		hasPartiallyStagedFiles: false,
		errors:                  set.New[error](),
		quiet:                   options.Quiet,
	}
}

/*

export const getInitialState = ({ quiet = false } = {}) => ({
  hasPartiallyStagedFiles: null,
  shouldBackup: null,
  errors: new Set([]),
  events: new EventEmitter(),
  output: [],
  quiet,
})

*/
