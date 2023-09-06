package hooks

import "github.com/spf13/cobra"

func Commands() []*cobra.Command {
	return []*cobra.Command{
		InstallCommand(),
		UninstallCommand(),
		AddCommand(),
		SetCommand(),
		InvokeCommand(),
	}
}
