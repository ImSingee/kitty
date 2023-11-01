package extregistry

import (
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
	// TODO set version to options

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
