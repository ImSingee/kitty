package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/spf13/cobra"

	"github.com/ImSingee/kitty/internal/lib/git"
)

type addOrSetOptions struct {
	name string // 'add' or 'set'
}

func SetCommand() *cobra.Command {
	o := &addOrSetOptions{name: "set"}

	return &cobra.Command{
		Use:    "set <hook> <cmd>",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.setHook(args[0], args[1])
		},
	}
}

func AddCommand() *cobra.Command {
	o := &addOrSetOptions{name: "add"}

	return &cobra.Command{
		Use:  "add <hook> [<cmd>]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			hook := args[0]
			cmd := ""
			if len(args) > 1 {
				cmd = args[1]
			}

			return o.addHook(hook, cmd)
		},
	}
}

func (o *addOrSetOptions) addHook(hook string, cmd string) error {
	fileName, cmd, err := o.convertHookToFile(hook, cmd)
	if err != nil {
		return err
	}

	return o.add(fileName, cmd)
}

func (o *addOrSetOptions) setHook(hook string, cmd string) error {
	fileName, cmd, err := o.convertHookToFile(hook, cmd)
	if err != nil {
		return err
	}

	return o.set(fileName, cmd)
}

func (o *addOrSetOptions) add(filename string, cmd string) error {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return o.set(filename, cmd)
		} else {
			return err
		}
	}

	l("updated %s", filename)

	if cmd == "" {
		return nil
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
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

func (o *addOrSetOptions) set(filename string, cmd string) error {
	userInputFile := filename
	filename, err := filepath.Abs(filename)
	if err != nil {
		return ee.Wrapf(err, "failed to get absolute path of %s", userInputFile)
	}

	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); err != nil {
		return ee.Wrapf(err, "can't create hook, %s directory doesn't exist (try running kitty install)", dir)
	}

	const fileHeader = `#!/usr/bin/env sh
. "$(dirname -- "$0")/_/kitty.sh"

`
	toWrite := fileHeader

	if cmd != "" {
		toWrite += cmd + "\n"
	}

	if err := os.WriteFile(filename, []byte(toWrite), 0755); err != nil {
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

func (o *addOrSetOptions) convertHookToFile(hook string, cmd string) (fileName string, newCmd string, err error) {
	if err := o.checkInstalled(); err != nil {
		return "", "", err
	}

	fileName = filepath.Join(".kitty", hook)

	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "@") {
		// use kitty extension
		cmd = "kitty " + cmd
	}

	return fileName, cmd, nil
}

func (o *addOrSetOptions) checkInstalled() error {
	if _, err := os.Stat(".git"); err != nil {
		return fmt.Errorf("this command must be run from the root of a git repository")
	}
	if _, err := os.Stat(".kitty"); err != nil {
		return fmt.Errorf("cannot found .kitty directory, please run 'kitty install' first")
	}

	return nil
}
