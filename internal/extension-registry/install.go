package extregistry

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/pp"
)

func (a *App) InstallUnknownVersionTo(version string, dst string) error {
	if a.TryGoInstall != nil {
		return goInstallTo(version, a.TryGoInstall, dst, true)
	}

	return ee.Errorf("unknown app version %s", version)
}

func (v *Version) InstallTo(dst string) error {
	err := tryInstall(
		dst, true,
		mr.Flats(
			v.downloaders(),
			v.fallbackDownloaders(),
		)...,
	)

	if err != nil {
		return ee.Wrapf(err, "cannot install app %s@%s", v.App.Name, v.Version)
	}

	return nil
}

func (v *Version) downloadTo(dst string) error {

	if u := v.Bin[osKey]; u != "" {
		return downloadFileTo(u, dst, 0755, true)
	}

	if v.GoInstall != "" {
		return goInstallTo(v.Version, v.GoInstall, dst, true)
	}

}

func (v *Version) downloaders() []Installer {
	return []Installer{
		&distInstaller{v.Bin},
	}
}

func (v *Version) fallbackDownloaders() []Installer {
	// return app's unknown downloaders
	return v.App.unknownDownloaders() // TODO
}

func tryInstall(dst string, showProgress bool, downloaders ...Installer) error {
	atLeastOneDownloaderAvailable := false
	for _, downloader := range downloaders {
		err := downloader.Install(dst, showProgress)
		if err == nil { // download success
			return nil
		}

		if ee.Is(err, ErrInstallerNotApplicable) {
			continue
		}

		atLeastOneDownloaderAvailable = true

		if showProgress {
			pp.ERedPrintln("Download fail:", err.Error())
		}
	}

	if !atLeastOneDownloaderAvailable {
		return ee.New("no ways to download")
	} else {
		return ee.New("all downloaders failed")
	}
}

func goInstallTo(version string, goOptions GoInstallOptions, dst string, showProgress bool) error {

	// pkg string

	d, err := mkdirFor(dst)
	if err != nil {
		return err
	}

	pkgWithVersion := pkg
	if !strings.Contains(pkg, "@") {
		// append version to pkg
		pkgWithVersion = pkg + "@" + normalizeGoVersion(version)
	}

	cmd := exec.Command("go", "install", pkgWithVersion)

	cmd.Env = []string{"GOBIN=" + d}
	cmd.Env = append(cmd.Env, os.Environ()...)

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

func goGetBinNameForPkg(pkg string) string {
	pkgWithoutVersion, _, _ := strings.Cut(pkg, "@")
	species := strings.Split(pkgWithoutVersion, "/")
	name := species[len(species)-1]

	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	return name
}

func mkdirFor(file string) (string, error) {
	d, err := filepath.Abs(file)
	if err != nil {
		return "", ee.Wrapf(err, "cannot get absolute path of %s", file)
	}

	d = filepath.Dir(d)

	err = os.MkdirAll(d, 0755)
	if err != nil {
		return "", ee.Wrapf(err, "cannot create directory %s", d)
	}

	return d, nil
}
