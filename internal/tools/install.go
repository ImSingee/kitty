package tools

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/pp"
	"github.com/spf13/cobra"
	"github.com/ysmood/gson"

	"github.com/ImSingee/kitty/internal/config"
	extregistry "github.com/ImSingee/kitty/internal/extension-registry"
	"github.com/ImSingee/kitty/internal/extension-registry/binkey"
	"github.com/ImSingee/kitty/internal/lib/git"
)

type installOptions struct {
	toInstall []string
}

func InstallCommand() *cobra.Command {
	o := &installOptions{}
	cmd := &cobra.Command{
		Use:     "install <app[@version]>",
		Aliases: []string{"add"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if isRoot, _ := git.IsRoot(""); !isRoot {
				return ee.Errorf("this command is only available in the root of a git repository")
			}

			// TODO lock to prevent concurrent install
			o.toInstall = args

			return o.install()
		},
	}

	return cmd
}

func (o *installOptions) install() error {
	// load tools from config
	currentTools, err := (&listOptions{}).getCurrentToolsMap()
	if err != nil {
		return ee.Wrap(err, "cannot load current tools")
	}
	// load tools from cli args
	for _, t := range o.toInstall {
		app, version, _ := strings.Cut(t, "@")
		if version == "" {
			version = "latest"
		}

		currentTools[app] = version
	}

	// load installed tools
	installedTools, err := (&listOptions{}).getInstalledTools()
	if err != nil {
		return ee.Wrap(err, "cannot load installed tools")
	}

	// diff
	var toInstall []string
	var toUpdate []string
	var toRemove []string

	for app, version := range currentTools {
		if version == "-" { // ignore
			continue
		}

		if installedVersion, ok := installedTools[app]; ok {
			if installedVersion != version {
				toUpdate = append(toUpdate, app)
			}
		} else {
			toInstall = append(toInstall, app)
		}
	}
	for app := range installedTools {
		if _, ok := currentTools[app]; !ok {
			toRemove = append(toRemove, app)
		}
	}

	// install needed (toInstall + toUpdate)
	toInstallOrUpdate := mr.Flats(toInstall, toUpdate)
	sort.Strings(toInstallOrUpdate)
	var installFailedApps []string
	results := make(map[string]*installResult, len(toInstallOrUpdate))
	for _, app := range toInstallOrUpdate {
		version := currentTools[app]

		pp.BluePrintln(">>> Install", app)

		o := &installOneOptions{app: app, version: version}

		result, err := o.install()
		if err != nil {
			pp.RedPrintln("ERROR:", err.Error())
			installFailedApps = append(installFailedApps, app)
			continue
		}
		results[app] = result
	}

	// remove unneeded (toRemove)
	for _, app := range toRemove {
		pp.BluePrintln(">>> Remove", app)

		// remove from .bin
		_ = os.Remove(filepath.Join(".kitty", ".bin", app))
	}

	// report error
	if len(installFailedApps) > 0 {
		return ee.Errorf("failed to install %s", strings.Join(installFailedApps, ", "))
	}

	// generate new tools info
	newToolsInfo := make(map[string]string, len(currentTools))
	for app, version := range currentTools {
		newToolsInfo[app] = version
	}
	for _, app := range toInstallOrUpdate {
		newToolsInfo[app] = results[app].version
	}

	// save tools info
	err = o.writeToolsInfo(newToolsInfo)
	if err != nil {
		return ee.Wrap(err, "cannot save tools info")
	}

	return nil
}

func (o *installOptions) writeToolsInfo(tools map[string]string) error {
	return config.PatchKittyConfig("", func(c map[string]gson.JSON) error {
		c["tools"] = gson.New(tools)
		return nil
	})
}

type installOneOptions struct {
	app     string
	version string
}

type installResult struct {
	version string
}

func (o *installOneOptions) install() (*installResult, error) {
	if o.app == "" {
		return nil, ee.Errorf("app name is empty")
	}
	if o.version == "" {
		o.version = "latest"
	}

	app, version, err := extregistry.GetAppVersion(o.app, o.version)
	if err != nil {
		return nil, err
	}

	result := &installResult{version: o.version}

	osKey := binkey.GetCurrentBinKey()

	// download to .bin/[system-key]/[name]@[version]
	rel := filepath.Join("."+string(osKey), app.Name+"@"+version.Version)
	dst := filepath.Join(".kitty", ".bin", rel)

	// TODO 安装到中央工具仓库（而不是 .bin 下）
	if version.InstallOptions != nil { // download for version
		err = version.InstallTo(dst)
		if err != nil {
			return nil, ee.Wrapf(err, "cannot download %s@%s", app.Name, version.Version)
		}
	} else { // download for unknown version
		err := version.InstallUnknownVersionTo(dst)
		if err != nil {
			return nil, ee.Wrapf(err, "cannot download %s@%s", app.Name, o.version)
		}
	}

	// Create soft link .bin/[name] -> .bin/[system-key]/[name]@[version]
	err = symlink(rel, filepath.Join(".kitty", ".bin", app.Name))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func symlink(from, to string) error {
	_ = os.Remove(to)

	err := os.Symlink(from, to)
	if err != nil {
		return ee.Wrapf(err, "cannot create sym link %s <- %s", to, from)
	}

	return nil
}
