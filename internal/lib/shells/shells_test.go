package shells

import (
	"testing"

	"github.com/ImSingee/tt"
)

func TestBuildCommandFromArgs(t *testing.T) {
	tt.AssertEqual(t, "ls -al", BuildCommandFromArgs("ls", "-al"))
	tt.AssertEqual(t, `sh -c 'ls -al'`, BuildCommandFromArgs("sh", "-c", "ls -al"))
}
