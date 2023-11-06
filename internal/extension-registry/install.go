package extregistry

import (
	"fmt"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"

	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/bininstaller"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/goinstaller"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

func (a *App) InstallUnknownVersionTo(version string, dst string) error {
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

func (v *Version) InstallTo(dst string) error {
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
	factoryIdl(&goinstaller.Factory{}),
}

var ErrNoWaysToDownload = fmt.Errorf("no ways to download")
