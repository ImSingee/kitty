package main

import (
	"bytes"
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	"github.com/ImSingee/kitty/internal/lib/git"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "embed"
)

const help = `Usage:
  kitty install
  kitty uninstall
  kitty add <hook-name> <cmd>
  kitty @extension ...
`

var debug bool
var extensions []*cobra.Command

func main() {
	app := &cobra.Command{
		Use:           "kitty [@extension]",
		Long:          help,
		Version:       getVersionString(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	app.AddCommand(
		&cobra.Command{
			Use:  "install",
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return install()
			},
		},
		&cobra.Command{
			Use:  "uninstall",
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return uninstall()
			},
		},
		&cobra.Command{
			Use:    "set <hook> <cmd>",
			Hidden: true,
			Args:   cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return set(args[0], args[1])
			},
		},
		&cobra.Command{
			Use:  "add <hook> <cmd>",
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return addHook(args[0], args[1])
			},
		},
		&cobra.Command{
			Use:    "@extension",
			Hidden: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				pp.Println("`@extension` is not a real command.\nIt just means you can use `kitty @xxx` to run some extension.\nFor example, if you want to run `lint-staged` extension, run `kitty @lint-staged`.")

				return nil
			},
		},
		&cobra.Command{
			Use:    "hook-invoke <hook-name> <version>",
			Args:   cobra.MinimumNArgs(2),
			Hidden: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				hookName := args[0]
				hookVersion := args[1]

				if hookVersion != "1" {
					fmt.Println(`echo "Your kitty version is too low, please upgrade"
exit 1`)
					return nil
				}

				fmt.Println(`export KITTY_VERSION=` + version)

				_ = hookName

				return nil
			},
		},
	)

	app.AddCommand(extensions...)

	// for global flags
	app.PersistentFlags().StringP("root", "R", "", "change command working directory")
	app.PersistentFlags().BoolVar(&debug, "debug", false, "print additional debug information")
	app.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if debug {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
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

//go:embed "kitty.sh"
var kittyDotShFile []byte

func install() error {
	dir := ".kitty" // TODO custom dir

	if os.Getenv("KITTY") == "0" {
		l("KITTY env variable is set to 0, skipping install")
		return nil
	}

	result := git.Run("rev-parse", "--show-toplevel")
	if result.ExitCode == -1 {
		l("git command not found, skipping install")
		return nil
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("not inside a git repository")
	}

	topLevel := string(bytes.TrimSpace(result.Output))

	// Ensure that cwd is git top level
	if _, err := os.Stat(".git"); err != nil {
		l(`Please go to the root of the git repository to run "kitty install"
> cd "` + topLevel + `"
> kitty install`)

		return ee.Phantom
	}

	// Start install

	// Create .kitty/_
	if err := os.MkdirAll(filepath.Join(dir, "_"), 0755); err != nil {
		l("Git hooks failed to install")
		return err
	}
	// Create .kitty/.gitignore
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("/_\n/.bin\n/.gitignore"), 0644); err != nil {
		l("Git hooks failed to install")
		return err
	}
	// Write .kitty/_/kitty.sh
	if err := os.WriteFile(filepath.Join(dir, "_", "kitty.sh"), kittyDotShFile, 0755); err != nil {
		l("Git hooks failed to install")
		return err
	}
	// Configure repo
	if err := git.Run("config", "core.hooksPath", dir).Err(); err != nil {
		l("Git hooks failed to install")
		return err
	}

	l("Git hooks installed")
	return nil
}

func uninstall() error {
	git.Run("config", "--unset", "core.hooksPath")
	return nil
}

func getKittyName(file string) string {
	if !strings.Contains(file, "/") && !strings.Contains(file, "\\") {
		return ".kitty/" + file
	}

	return file
}

// Create a hook file if it doesn't exist or overwrite it
func set(file string, cmd string) error {
	file = getKittyName(file)
	userInputFile := file
	file, err := filepath.Abs(file)
	if err != nil {
		return ee.Wrapf(err, "failed to get absolute path of %s", userInputFile)
	}

	dir := filepath.Dir(file)
	if _, err := os.Stat(dir); err != nil {
		return ee.Wrapf(err, "can't create hook, %s directory doesn't exist (try running kitty install)", dir)
	}

	const fileHeader = `#!/usr/bin/env sh
. "$(dirname -- "$0")/_/kitty.sh"

`
	toWrite := fileHeader + cmd + "\n"

	if err := os.WriteFile(file, []byte(toWrite), 0755); err != nil {
		return ee.Wrapf(err, "failed to write hook file %s", userInputFile)
	}

	l("created %s", userInputFile)

	if runtime.GOOS == "windows" {
		l(
			`Due to a limitation on Windows systems, the executable bit of the file cannot be set without using git.
To fix this, the file ${file} has been automatically moved to the staging environment and the executable bit has been set using git.
Note that, if you remove the file from the staging environment, the executable bit will be removed.
You can add the file back to the staging environment and include the executable bit using the command 'git update-index -add --chmod=+x ${file}'.
If you have already committed the file, you can add the executable bit using 'git update-index --chmod=+x ${file}'.
You will have to commit the file to have git keep track of the executable bit.`)

		git.Run("update-index", "--add", "--chmod=+x", userInputFile)
	}

	return nil
}

func checkEnv() error {
	if _, err := os.Stat(".git"); err != nil {
		return fmt.Errorf("this command must be run from the root of a git repository")
	}
	if _, err := os.Stat(".kitty"); err != nil {
		return fmt.Errorf("cannot found .kitty directory, please run 'kitty install' first")
	}

	return nil
}

func addHook(name string, cmd string) error {
	if err := checkEnv(); err != nil {
		return err
	}

	fileName := filepath.Join(".kitty", name)

	if strings.HasPrefix(cmd, "@") {
		// use kitty extension
		cmd = "kitty " + cmd
	}

	return add(fileName, cmd)
}

// Create a hook if it doesn't exist or append command to it
func add(file string, cmd string) error {
	file = getKittyName(file)

	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return set(file, cmd)
		} else {
			return err
		}
	}

	l("updated %s", file)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(cmd + "\n"))
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
