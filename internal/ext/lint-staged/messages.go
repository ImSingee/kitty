package lintstaged

import (
	"fmt"

	"github.com/ImSingee/go-ex/pp"
)

const (
	info    = arrowRight
	warning = "⚠"
	x       = "✗"
	yes     = "✔"

	arrowRight = "→"
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

	return fmt.Sprintf("%s Skipping backup because %s.", warning, reason)
}

var gray = pp.GetColor(38, 5, 240)

func symGray(s string) string {
	return pp.ColorString(gray, s).GetForStdout()
}
