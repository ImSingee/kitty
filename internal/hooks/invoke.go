package hooks

import (
	"github.com/ImSingee/go-ex/pp"
	"github.com/spf13/cobra"

	"github.com/ImSingee/kitty/internal/lib/shells"
)

type invokeOptions struct {
	hookName    string
	hookVersion string
}

func InvokeCommand() *cobra.Command {
	o := &invokeOptions{}

	return &cobra.Command{
		Use:    "hook-invoke <hook-name> <version>",
		Args:   cobra.MinimumNArgs(2),
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			o.hookName = args[0]
			o.hookVersion = args[1]

			o.invokeWrapper()
		},
	}
}

func (o *invokeOptions) invokeWrapper() {
	output, success := o.invoke()
	if output != "" {
		pp.Println(`echo ` + shells.Quote(output))
	}
	if !success {
		pp.Println("exit 1")
	}
}

func (o *invokeOptions) invoke() (output string, success bool) {
	if o.hookVersion != "1" {
		return "Your kitty version is too low, please upgrade", false
	}

	//pp.Println(`export KITTY_VERSION=` + version) // TODO
	return "", true
}
