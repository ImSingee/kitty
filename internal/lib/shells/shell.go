package shells

import (
	"fmt"

	"github.com/alessio/shellescape"
	"github.com/google/shlex"
)

func Quote(arg string) string {
	return shellescape.Quote(arg)
}

func Join(cmdAndArgs []string) string {
	return shellescape.QuoteCommand(cmdAndArgs)
}

func Split(cmd string) ([]string, error) {
	return shlex.Split(cmd)
}

func MustSplit(cmd string) []string {
	a, err := Split(cmd)
	if err != nil {
		panic("Cannot split command `" + cmd + "`: " + err.Error())
	}
	return a
}

// BuildCommandFromArgs 如果 args 为空则使用 name 作为整个 command，否则将 args 作为参数并 quote
func BuildCommandFromArgs(name string, args ...string) string {
	if len(args) == 0 {
		return name
	} else {
		return fmt.Sprintf("%s %s", name, shellescape.QuoteCommand(args))
	}
}
