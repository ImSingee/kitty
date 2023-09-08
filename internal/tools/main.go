package tools

import "github.com/spf13/cobra"

func Commands() []*cobra.Command {
	cmd := &cobra.Command{
		Use:     "tools",
		Aliases: []string{"tool", "t"},
	}

	// kitty tools install 自动安装尚未安装/版本不匹配的 tools
	// kitty tools install xxx 安装指定的 tools（改成做 diff 安装，类似 pnpm）
	// kitty install 自动执行 kitty tools install

	cmd.AddCommand(
		InstallCommand(),
	)

	return []*cobra.Command{cmd}
}
