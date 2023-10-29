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
	urlObject := o[string(GetCurrentBinKey())].Val()

	url := ""
	switch urlObject := urlObject.(type) {
	case string:
		url = urlObject
	case map[string]any:
		if u, ok := urlObject["url"].(string); ok {
			url = u
		}
	}

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
