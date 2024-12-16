package goinstaller

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
	"github.com/ImSingee/semver"

	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
	erutils "github.com/ImSingee/kitty/internal/extension-registry/utils"
)

type Factory struct{}

func (Factory) Key() string {
	return "go-install"
}

func (Factory) GetInstaller(o eroptions.AnyOptions) (installer.Installer, error) {
	eroptions.RenameDollarKey(o, "pkg", false)

	url, _ := o["pkg"].Val().(string)

	if url == "" {
		return nil, installer.ErrInstallerNotApplicable
	}

	return &Installer{url}, nil
}

type Installer struct {
	pkg string
}

func (i *Installer) Install(o *installer.InstallOptions) error {
	pkg := i.pkg // TODO template render

	return goInstallTo(pkg, o.Version, o.To, o.ShowProgress)
}

func goInstallTo(pkg, version string, dst string, showProgress bool) error {
	d, err := erutils.MkdirFor(dst)
	if err != nil {
		return err
	}

	pkgWithVersion := pkg
	if !strings.Contains(pkg, "@") {
		// append version to pkg
		pkgWithVersion = pkg + "@" + normalizeGoVersion(version)
	}

	cmd := exec.Command("go", "install", pkgWithVersion)

	cmd.Env = append(copyEnvWithoutGOBIN(), "GOBIN="+d)

	if showProgress {
		pp.Println("go install", pkgWithVersion, "...")

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	err = cmd.Run()
	if err != nil {
		return ee.Wrapf(err, "cannot install %s", pkgWithVersion)
	}

	// rename bin
	goBinName := filepath.Join(d, goGetBinNameForPkg(pkg))
	err = os.Rename(goBinName, dst)
	if err != nil {
		return ee.Wrapf(err, "cannot rename %s to %s", goBinName, dst)
	}

	return nil
}

func normalizeGoVersion(v string) string {
	if v == "" {
		return "latest"
	}

	if version, err := semver.NewVersion(v); err == nil {
		return "v" + version.String()
	}

	return v
}

func goGetBinNameForPkg(pkg string) string {
	pkgWithoutVersion, _, _ := strings.Cut(pkg, "@")
	species := strings.Split(pkgWithoutVersion, "/")
	name := species[len(species)-1]

	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	return name
}

func copyEnvWithoutGOBIN() []string {
	env := os.Environ()
	copied := make([]string, 0, len(env))

	for _, e := range env {
		if strings.HasPrefix(e, "GOBIN=") {
			continue
		}

		copied = append(copied, e)
	}

	return copied
}
