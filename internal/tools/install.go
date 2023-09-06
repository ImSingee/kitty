package tools

import (
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	extregistry "github.com/ImSingee/kitty/internal/extension-registry"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type installOptions struct {
	app     string
	version string

	noPersistence bool
}

func InstallCommand() *cobra.Command {
	opts := &installOptions{}
	cmd := &cobra.Command{
		Use:  "install <app[@version]>",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.noPersistence {
				// TODO check is repo root
			}

			failApps := []string{}

			for _, arg := range args {
				o := *opts

				pp.BluePrintln(">>> Install", arg)
				o.app, o.version, _ = strings.Cut(arg, "@")
				err := o.install()
				if err != nil {
					pp.RedPrintln("ERROR:", err.Error())
					failApps = append(failApps, arg)
				}
			}

			if len(failApps) > 0 {
				return ee.Errorf("failed to install %s", strings.Join(failApps, ", "))
			}

			return nil
		},
	}

	return cmd
}

func (o *installOptions) install() error {
	if o.app == "" {
		return ee.Errorf("app name is empty")
	}
	if o.version == "" {
		o.version = "latest"
	}

	app, version, err := extregistry.GetAppVersion(o.app, o.version)
	if err != nil {
		return err
	}

	osKey := extregistry.GetCurrentBinKey()

	// download to .bin/[system-key]/[name]@[version]
	var rel string
	if version != nil { // download for version
		rel = filepath.Join("."+string(osKey), app.Name+"@"+version.Version)
		dst := filepath.Join(".kitty", ".bin", rel)

		err = version.DownloadTo(dst)
		if err != nil {
			return ee.Wrapf(err, "cannot download %s@%s", app.Name, version.Version)
		}
	} else { // download for unknown version
		rel = filepath.Join("."+string(osKey), app.Name+"@"+o.version)
		dst := filepath.Join(".kitty", ".bin", rel)

		err := app.DownloadUnknownVersionTo(o.version, dst)
		if err != nil {
			return ee.Wrapf(err, "cannot download %s@%s", app.Name, o.version)
		}
	}

	// Create soft link .bin/[name] -> .bin/[system-key]/[name]@[version]
	err = symlink(rel, filepath.Join(".kitty", ".bin", app.Name))
	if err != nil {
		return err
	}

	return nil
}

func symlink(from, to string) error {
	_ = os.Remove(to)

	err := os.Symlink(from, to)
	if err != nil {
		return ee.Wrapf(err, "cannot create sym link %s <- %s", to, from)
	}

	return nil
}
