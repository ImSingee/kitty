package bininstaller

import (
	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
	erutils "github.com/ImSingee/kitty/internal/extension-registry/utils"
)

type Factory struct{}

func (Factory) Key() string {
	return "bin"
}

func (Factory) GetInstaller(o eroptions.AnyOptions) (installer.Installer, error) {
	object := eroptions.Get(o, string(GetCurrentBinKey()))
	maybeUrl := eroptions.RenameDollarKey(object, "url", false)["url"].Val()

	url, _ := maybeUrl.(string)

	if url == "" {
		return nil, installer.ErrInstallerNotApplicable
	}

	return &Installer{url}, nil
}

type Installer struct {
	url string
}

func (i *Installer) Install(o *installer.InstallOptions) error {
	url := i.url // TODO template render

	return erutils.DownloadFileTo(url, o.To, 0755, o.ShowProgress)
}
