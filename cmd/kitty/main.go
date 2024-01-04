package main

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	"github.com/spf13/cobra"

	"github.com/ImSingee/kitty/internal/config"
	"github.com/ImSingee/kitty/internal/ext/format"
	lintstaged "github.com/ImSingee/kitty/internal/ext/lint-staged"
	"github.com/ImSingee/kitty/internal/hooks"
	"github.com/ImSingee/kitty/internal/lib/git"
	"github.com/ImSingee/kitty/internal/lib/xlog"
	"github.com/ImSingee/kitty/internal/tools"
	"github.com/ImSingee/kitty/internal/version"

	_ "embed"
)

const help = `Usage:
  kitty install
  kitty add <hook-name> <cmd>
  kitty tools install <tool-name>
  kitty @extension ...
`

func main() {
	app := &cobra.Command{
		Use:           "kitty [@extension]",
		Long:          help,
		Version:       version.GetVersionString(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	app.AddCommand(hooks.Commands()...)
	app.AddCommand(tools.Commands()...)

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
	// load internal extensions
	app.AddCommand(lintstaged.Commands()...)
	app.AddCommand(format.Commands()...)

	// for extension
	app.TraverseChildren = true
	app.Flags().SetInterspersed(false)
	app.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || len(args[0]) <= 1 || !strings.HasPrefix(args[0], "@") {
			return cmd.Help()
		}

		extensionName := args[0][1:]
		extensionArgs := args[1:]

		return runExtension(extensionName, extensionArgs)
	}

	// for global flags
	app.PersistentFlags().SortFlags = false
	app.PersistentFlags().StringP("root", "R", "", "change command working directory")
	app.PersistentFlags().BoolVar(&config.Debug, "debug", false, "print additional debug information")
	app.PersistentFlags().BoolP("quiet", "q", false, "quiet mode (hide any output)")

	// pre-run
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

	// run!
	err := app.Execute()
	if err != nil {
		if !ee.Is(err, ee.Phantom) {
			pp.ERedPrintln("Error:", err.Error())
		}

		os.Exit(1)
	}
}

func runExtension(name string, args []string) error {
	root, err := git.GetRoot("")
	if err != nil {
		return ee.Wrap(err, "cannot get git root")
	}

	// TODO verify tools (check exist, check version)
	// TODO call tools.GetTool (like below, refactor later)
	// run apps
	if appBin, err := exec.LookPath(filepath.Join(root, ".kitty", ".bin", name)); err == nil {
		// bin extension

		cmd := exec.Command(appBin, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Env = append([]string{
			"KITTY_GIT_ROOT=" + root,
		}, os.Environ()...)

		return cmd.Run()
	}

	return ee.Errorf("unknown extension `%s`", name)
}
