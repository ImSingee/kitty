package bininstaller

import (
	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

type Factory struct{}

func (Factory) Key() string {
	return "bin"
}

func (Factory) GetInstaller(o eroptions.AnyOptions) (installer.Installer, error) {
	return &Installer{}, nil
}

type Installer struct {
}

func (*Installer) Install(o *installer.InstallOptions) error {
	return installer.ErrSkip // TODO
}
