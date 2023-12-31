package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/spf13/cobra"

	"github.com/ImSingee/kitty/internal/lib/git"

	_ "embed"
)

type installOptions struct {
	hideSuccessMessageIfNotFirstInstall bool // only print success message in first install
	doNotInstallTools                   bool
	generateEnvRc                       bool
}

func InstallCommand() *cobra.Command {
	o := &installOptions{}

	fromDirEnv := false

	cmd := &cobra.Command{
		Use:  "install",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromDirEnv {
				o.hideSuccessMessageIfNotFirstInstall = true
			}

			return o.install()
		},
	}

	flags := cmd.Flags()

	flags.SortFlags = false
	flags.BoolVar(&o.generateEnvRc, "direnv", false, "generate .envrc file")
	flags.BoolVar(&fromDirEnv, "from-direnv", false, "")
	_ = flags.MarkHidden("from-direnv")
	flags.BoolVar(&o.doNotInstallTools, "no-tools", false, "do not install tools")

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

	if !kittyShExists || !o.hideSuccessMessageIfNotFirstInstall {
		l("Git hooks installed")
	}

	if !o.doNotInstallTools {
		if err := o.installTools(); err != nil {
			return ee.Wrap(err, "cannot install tools")
		}
	}

	if o.generateEnvRc {
		err := o.writeEnvRcFile()
		if err != nil {
			return err
		}
	}

	return nil
}

//go:embed "envrc"
var dotEnvRcFile []byte

func (o *installOptions) writeEnvRcFile() error {
	_, err := os.Stat(".envrc")
	if err == nil {
		l(`.envrc file already exists, please write following content to your .envrc file manually:` + "\n" + strings.TrimSpace(string(dotEnvRcFile)))
		return fmt.Errorf(".envrc file already exists")
	}
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to access .envrc file: %w", err)
		}
	}

	return os.WriteFile(".envrc", dotEnvRcFile, 0755)
}

func (o *installOptions) installTools() error {
	// run `kitty tools install` in a subprocess

	self, err := os.Executable()
	if err != nil {
		self = "/proc/self/exe" // linux only
	}

	cmd := exec.Command(self, "tools", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
