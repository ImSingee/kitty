package extregistry

import (
	"fmt"
	"log/slog"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"

	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/bininstaller"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/bininstaller/tarinstaller"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/bininstaller/zipinstaller"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/goinstaller"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

func (a *App) installUnknownVersionTo(version string, dst string) error {
	slog.Debug("app installUnknownVersionTo", "version", version, "dst", dst)
	o := &installer.InstallOptions{
		Version:      version,
		To:           dst,
		ShowProgress: true,
	}

	err := tryInstall(a.InstallOptions, o)
	if err != nil {
		if ee.Is(err, ErrNoWaysToDownload) {
			return ee.Errorf("unknown app version %s", version)
		}

		return ee.Wrapf(err, "cannot install app %s@%s", a.Name, version)
	}

	return nil
}

func (v *Version) InstallUnknownVersionTo(dst string) error {
	slog.Debug("version installUnknownVersionTo", "dst", dst)
	return v.App.installUnknownVersionTo(v.Version, dst)
}

func (v *Version) InstallTo(dst string) error {
	slog.Debug("version InstallTo", "version", v, "dst", dst)
	o := &installer.InstallOptions{
		Version:      v.Version,
		To:           dst,
		ShowProgress: true,
	}

	err := tryInstall(v.InstallOptions, o)

	if err != nil {
		return ee.Wrapf(err, "cannot install app %s@%s", v.App.Name, v.Version)
	}

	return nil
}

func tryInstall(o0 eroptions.AnyOptions, o1 *installer.InstallOptions) error {
	atLeastOneDownloaderAvailable := false
	for _, install := range installers {
		err := install(o0, o1)
		if err == nil { // download success
			return nil
		}

		if ee.Is(err, installer.ErrSkip) {
			continue
		}

		atLeastOneDownloaderAvailable = true

		if o1.ShowProgress {
			pp.ERedPrintln("Download fail:", err.Error())
		}
	}

	if !atLeastOneDownloaderAvailable {
		return ErrNoWaysToDownload
	} else {
		return ee.New("all downloaders failed")
	}
}

// idl is a lazy Factory + Installer wrapper
// it will return ErrSkip if the installer is not applicable
type idl func(o0 eroptions.AnyOptions, o1 *installer.InstallOptions) error

func factoryIdl(factory installer.Factory) idl {
	return func(o0 eroptions.AnyOptions, o1 *installer.InstallOptions) error {
		key := factory.Key()
		o2 := o0
		if key != "" {
			o2 = eroptions.Get(o0, key)
		}
		if o2 == nil {
			return installer.ErrSkip
		}

		installer1, err := factory.GetInstaller(o2)
		if err != nil {
			return installer.ErrSkip
		}

		return installer1.Install(o1)
	}
}

var installers = []idl{
	factoryIdl(&bininstaller.Factory{}),
	factoryIdl(&zipinstaller.Factory{}),
	factoryIdl(&tarinstaller.Factory{}),
	factoryIdl(&goinstaller.Factory{}),
}

var ErrNoWaysToDownload = fmt.Errorf("no ways to download")
