package zipinstaller

import (
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"

	"github.com/ImSingee/kitty/internal/extension-registry/binkey"
	"github.com/ImSingee/kitty/internal/extension-registry/installer"
	"github.com/ImSingee/kitty/internal/extension-registry/installer/tmpl"
	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
	erutils "github.com/ImSingee/kitty/internal/extension-registry/utils"
)

type Factory struct{}

func (Factory) Key() string {
	return "dist-zip"
}

func (Factory) GetInstaller(o eroptions.AnyOptions) (installer.Installer, error) {
	object := eroptions.Get(o, string(binkey.GetCurrentBinKey()))
	object = eroptions.RenameDollarKey(object, "url", false)

	if eroptions.Exists(o, "default") {
		defaultOptions := eroptions.Get(o, "default")
		defaultOptions = eroptions.RenameDollarKey(defaultOptions, "url", false)

		object = eroptions.Merge(object, defaultOptions, false)
	}

	url, _ := object["url"].Val().(string)
	bin, _ := object["bin"].Val().(string)

	if url == "" || bin == "" {
		return nil, installer.ErrInstallerNotApplicable
	}

	return &Installer{url, bin}, nil
}

type Installer struct {
	url string
	bin string
}

func (i *Installer) Install(o *installer.InstallOptions) error {
	url, err := tmpl.Render(i.url, o)
	if err != nil {
		return ee.Wrap(err, "invalid url")
	}
	bin, err := tmpl.Render(i.bin, o)
	if err != nil {
		return ee.Wrap(err, "invalid bin")
	}

	filenameWithoutExt, ext, err := getFilenamePartsOfUrl(url)
	if err != nil {
		return err
	}
	filename := filenameWithoutExt + ext

	zipDownloadedTo, err := erutils.DownloadFileToTemp(url, "kdl-*-"+filename, o.ShowProgress)
	if err != nil {
		return ee.Wrap(err, "cannot download zip file")
	}

	if o.ShowProgress {
		pp.Println("zip file downloaded to:", zipDownloadedTo)
	}

	unzipToDir, err := os.MkdirTemp("", "kunzip-*")
	if err != nil {
		return ee.Wrap(err, "cannot create temp dir")
	}

	if o.ShowProgress {
		pp.Println("Extract ...")
	}
	err = erutils.Unzip(zipDownloadedTo, unzipToDir)
	if err != nil {
		return ee.Wrap(err, "cannot unzip file")
	}

	distFrom := path.Join(unzipToDir, bin)
	if _, err := os.Lstat(distFrom); err != nil {
		return ee.Wrapf(err, "cannot find bin file %s in zip", bin)
	}

	_, _ = erutils.MkdirFor(o.To)
	err = erutils.Rename(distFrom, o.To)
	if err != nil {
		return ee.Wrapf(err, "failed to install bin file %s from zip", bin)
	}

	return nil
}

func getFilenamePartsOfUrl(u string) (name string, ext string, err error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", "", ee.Wrapf(err, "url %s is invalid", u)
	}

	paths := strings.Split(parsed.Path, "/")
	pathLastPart := paths[len(paths)-1]

	dotIndex := strings.Index(pathLastPart, ".")

	if dotIndex == -1 {
		return pathLastPart, "", nil
	} else {
		return pathLastPart[:dotIndex], pathLastPart[dotIndex:], nil
	}
}
