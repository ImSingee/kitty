package lintstaged

import "github.com/ImSingee/go-ex/pp"

const (
	info    = arrowRight
	warning = "⚠"

	arrowRight = "→"
)

const (
	NO_STAGED_FILES = info + " No staged files found."
)

func skippingBackup(hasInitialCommit bool, diff string) string {
	var reason string
	switch {
	case diff != "":
		reason = "`--diff` was used"
	case hasInitialCommit:
		reason = "`--stash=false` was used"
	default:
		reason = "there’s no initial commit yet"
	}

	return pp.YellowString("%s Skipping backup because %s.\n", warning, reason).GetForStdout()
}
