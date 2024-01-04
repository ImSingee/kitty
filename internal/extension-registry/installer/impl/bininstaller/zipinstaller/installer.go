package zipinstaller

import (
	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/impl/bininstaller/indirect"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

type Factory struct{}

func (Factory) Key() string {
	return "dist-zip"
}

func (Factory) GetInstaller(o eroptions.AnyOptions) (installer.Installer, error) {
	return indirect.GetInstaller("zip", o)
}
