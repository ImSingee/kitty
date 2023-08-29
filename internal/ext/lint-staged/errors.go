package lintstaged

import "fmt"

var (
	ErrApplyEmptyCommit       = fmt.Errorf("apply empty commit error")
	ErrConfigNotFound         = fmt.Errorf("configuration could not be found")
	ErrConfigFormat           = fmt.Errorf("configuration should be an object or a function") // TODO
	ErrConfigEmpty            = fmt.Errorf("configuration should not be empty")
	ErrGetBackupStash         = fmt.Errorf("get backup stash error")
	ErrGetStagedFiles         = fmt.Errorf("get staged files error")
	ErrGitRepo                = fmt.Errorf("git repo error")
	ErrHideUnstagedChanges    = fmt.Errorf("hide unstaged changes error")
	ErrInvalidOptions         = fmt.Errorf("invalid options")
	ErrRestoreMergeStatus     = fmt.Errorf("restore merge status error")
	ErrRestoreOriginalState   = fmt.Errorf("restore original state error")
	ErrRestoreUnstagedChanges = fmt.Errorf("restore unstaged changes error")
)
