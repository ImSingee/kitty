package hooks

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/kitty/internal/lib/git"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type installOptions struct {
	notFirst bool // only print success message in first install
}

func InstallCommand() *cobra.Command {
	o := &installOptions{}

	cmd := &cobra.Command{
		Use:  "install",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.install()
		},
	}

	flags := cmd.Flags()

	flags.SortFlags = false
	flags.BoolVar(&o.notFirst, "not-first", false, "")
	_ = flags.MarkHidden("not-first")

	return cmd
}

//go:embed "kitty.sh"
var kittyDotShFile []byte

func (o *installOptions) install() error {
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
	_, kittyShStatErr := os.Stat(filepath.Join(dir, "_", "kitty.sh"))
	kittyShExists := kittyShStatErr == nil

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

	if !kittyShExists || !o.notFirst {
		l("Git hooks installed")
	}

	return nil
}
