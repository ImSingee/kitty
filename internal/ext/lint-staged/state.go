package lintstaged

import "github.com/ImSingee/go-ex/set"

type State struct {
	hasPartiallyStagedFiles bool
	deletedFiles            []string
	shouldBackup            bool
	errors                  *set.Set[error]
	events                  *struct{} // TODO
	output                  []string
	quiet                   bool
}

func getInitialState(options *Options) *State {
	return &State{
		hasPartiallyStagedFiles: false,
		errors:                  set.New[error](),
		events:                  nil, // TODO
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
