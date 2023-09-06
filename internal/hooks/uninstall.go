package hooks

import (
	"github.com/ImSingee/kitty/internal/lib/git"
	"github.com/spf13/cobra"
)

type uninstallOptions struct{}

func UninstallCommand() *cobra.Command {
	o := &uninstallOptions{}
	return &cobra.Command{
		Use:  "uninstall",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.uninstall()
		},
	}
}

func (o *uninstallOptions) uninstall() error {
	git.Run("config", "--unset", "core.hooksPath")
	return nil
}
