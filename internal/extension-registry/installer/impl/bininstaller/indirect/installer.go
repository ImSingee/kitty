package indirect

import (
	"log/slog"
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

func GetInstaller(t string, o eroptions.AnyOptions) (installer.Installer, error) {
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

	return &Installer{t, url, bin}, nil
}

type Installer struct {
	t   string
	url string
	bin string
}

func (i *Installer) Install(o *installer.InstallOptions) error {
	t := i.t

	slog.Debug("tar/zip installer", "t", t, "url", i.url, "bin", i.bin)

	url, err := tmpl.Render(i.url, o)
	if err != nil {
		return ee.Wrap(err, "invalid url")
	}
	bin, err := tmpl.Render(i.bin, o)
	if err != nil {
		return ee.Wrap(err, "invalid bin")
	}

	slog.Debug("tar/zip installer (rendered)", "t", t, "url", url, "bin", bin)

	filenameWithoutExt, ext, err := getFilenamePartsOfUrl(url)
	if err != nil {
		return err
	}
	filename := filenameWithoutExt + ext

	slog.Debug("tar/zip installer", "filename", filename)

	zipDownloadedTo, err := erutils.DownloadFileToTemp(url, "kdl-*-"+filename, o.ShowProgress)
	if err != nil {
		return ee.Wrapf(err, "cannot download %s file", t)
	}

	if o.ShowProgress {
		pp.Printf("%s file downloaded to: %s\n", t, zipDownloadedTo)
	}

	unzipToDir, err := os.MkdirTemp("", "kunzip-*")
	if err != nil {
		return ee.Wrap(err, "cannot create temp dir")
	}
	slog.Debug("tar/zip installer untar/unzip to", "target", unzipToDir)

	if o.ShowProgress {
		pp.Println("Extract ...")
	}

	switch t {
	case "zip":
		err = erutils.Unzip(zipDownloadedTo, unzipToDir)
	case "tar":
		err = erutils.Untar(zipDownloadedTo, unzipToDir)
	default:
		panic("zip/tar installer got unknown t " + t)
	}

	if err != nil {
		return ee.Wrapf(err, "cannot un%s file", t)
	}

	distFrom := path.Join(unzipToDir, bin)
	if _, err := os.Lstat(distFrom); err != nil {
		return ee.Wrapf(err, "cannot find bin file %s in %s", bin, t)
	}

	_, _ = erutils.MkdirFor(o.To)
	err = erutils.Rename(distFrom, o.To)
	if err != nil {
		return ee.Wrapf(err, "failed to install bin file %s from %s", bin, t)
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
