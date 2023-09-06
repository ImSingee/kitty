package main

import (
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	"github.com/ImSingee/kitty/internal/config"
	"github.com/ImSingee/kitty/internal/hooks"
	"github.com/ImSingee/kitty/internal/lib/xlog"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
	"strings"

	_ "embed"
)

const help = `Usage:
  kitty install
  kitty uninstall
  kitty add <hook-name> <cmd>
  kitty @extension ...
`

var extensions []*cobra.Command

func main() {
	app := &cobra.Command{
		Use:           "kitty [@extension]",
		Long:          help,
		Version:       getVersionString(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	app.AddCommand(hooks.Commands()...)

	app.AddCommand(
		&cobra.Command{
			Use:    "@extension",
			Hidden: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				pp.Println("`@extension` is not a real command.\nIt just means you can use `kitty @xxx` to run some extension.\nFor example, if you want to run `lint-staged` extension, run `kitty @lint-staged`.")

				return nil
			},
		},
	)

	app.AddCommand(extensions...)

	// for global flags
	app.PersistentFlags().SortFlags = false
	app.PersistentFlags().StringP("root", "R", "", "change command working directory")
	app.PersistentFlags().BoolVar(&config.Debug, "debug", false, "print additional debug information")
	app.PersistentFlags().BoolP("quiet", "q", false, "quiet mode (hide any output)")
	app.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			if null, _ := os.Open("/dev/null"); null != nil {
				os.Stdout = null
				os.Stderr = null
			}

			pp.Stdout.ChangeWriter(io.Discard)
			pp.Stderr.ChangeWriter(io.Discard)

			slog.SetDefault(xlog.DisabledLogger)
		}

		if !quiet { // setup logger
			if config.Debug {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}
		}

		if root, _ := app.PersistentFlags().GetString("root"); root != "" {
			slog.Debug("Change working directory", "root", root)
			err := os.Chdir(root)
			if err != nil {
				return ee.Wrapf(err, "cannot change working directory to %s", root)
			}
		}

		return nil
	}

	// for extension
	app.TraverseChildren = true
	app.Flags().SetInterspersed(false)
	app.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || len(args[0]) <= 1 || !strings.HasPrefix(args[0], "@") {
			return cmd.Help()
		}

		extensionName := args[0][1:]
		//extensionArgs := args[1:]
		// TODO real run extensions

		return ee.Errorf("unknown extension `%s`", extensionName)
	}

	// run!
	err := app.Execute()
	if err != nil {
		if !ee.Is(err, ee.Phantom) {
			l("Error: %v", err)
		}

		os.Exit(1)
	}
}

func l(msg string, args ...any) {
	s := msg
	if len(args) != 0 {
		s = fmt.Sprintf(msg, args...)
	}

	_, _ = os.Stderr.Write([]byte("kitty - " + strings.TrimSpace(s) + "\n"))
}
