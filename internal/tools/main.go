package tools

import "github.com/spf13/cobra"

func Commands() []*cobra.Command {
	cmd := &cobra.Command{
		Use:     "tools",
		Aliases: []string{"tool"},
	}

	cmd.AddCommand(
		InstallCommand(),
	)

	return []*cobra.Command{cmd}
}
