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
	"github.com/ImSingee/semver"
	"github.com/spf13/cobra"
	"github.com/ysmood/gson"

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

		err := mayUseAnotherKitty()
		if err != nil {
			return err
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

func mayUseAnotherKitty() error {
	wd, err := os.Getwd()
	if err != nil {
		return ee.Wrap(err, "cannot get working directory")
	}

	gitRoot, err := git.GetRoot(wd)
	if err != nil {
		return nil // cannot found git root
	}

	c, err := config.GetKittyConfig(gitRoot)
	if err != nil {
		if config.IsNotExist(err) {
			return nil // no kitty config
		} else {
			return ee.Wrap(err, "cannot get kitty config")
		}
	}

	requiredVersion := parseRequiredKittyVersion(c)
	if requiredVersion == "" {
		return nil // no required version
	}

	// TODO auto download new kitty version
	if requiredVersion[0] == '=' {
		sv, err := semver.NewVersion(requiredVersion[1:])
		if err != nil {
			return ee.Wrapf(err, "invalid kitty required version %s", requiredVersion)
		}

		if !sv.Equal(version.Semver()) {
			pp.Println("Please use kitty ", requiredVersion[1:], "to run this command")
			pp.Println("Visit https://github.com/ImSingee/kitty/releases/tag/v" + requiredVersion[1:] + " to download")
			return ee.Phantom
		}
	} else {
		sv, err := semver.NewVersion(strings.TrimPrefix(requiredVersion, ">"))
		if err != nil {
			return ee.Wrapf(err, "invalid kitty required version %s", requiredVersion)
		}

		if version.LessThan(sv) {
			pp.Println("Please use kitty ", requiredVersion, "or later to run this command")
			pp.Println("Visit https://github.com/ImSingee/kitty/releases to download")
			return ee.Phantom
		}
	}

	return nil
}

func parseRequiredKittyVersion(c map[string]gson.JSON) string {
	if c == nil {
		return ""
	}

	kitty, kittyExists := c["kitty"]
	if !kittyExists {
		return ""
	}

	switch kittyVal := kitty.Val().(type) {
	case string:
		return kittyVal
	case map[string]any:
		if v, ok := kittyVal["version"].(string); ok {
			return v
		}

		return ""
	default:
		return "" // unknown kitty config type
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
