package extregistry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"

	"github.com/ImSingee/kitty/internal/config"
)

// GetAppVersion will get the app and version from registry
//
// the returned version may be empty
func GetAppVersion(app string, version string) (*App, *Version, error) {
	a, err := GetApp(app)
	if err != nil {
		return nil, nil, ee.Wrapf(err, "cannot get app %s from registry", app)
	}

	if realVersion := a.Tags[version]; realVersion != "" {
		version = realVersion
	}

	v := a.Versions[version]

	if v != nil {
		v.Version = version
	}

	return a, v, nil
}

func GetApp(app string) (*App, error) {
	r, err := GetRegistry()
	if err != nil {
		return nil, ee.Wrap(err, "cannot get registry")
	}

	urlForApp, err := url.JoinPath(r, app, "manifest.json")
	if err != nil {
		return nil, ee.Wrapf(err, "cannot generate manifest url for app (registry = %s)", r)
	}

	a, err := getAppFromUrl(app, urlForApp)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func getAppFromUrl(name string, url string) (*App, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, ee.Wrapf(err, "cannot get app manifest from %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ee.Errorf("cannot get app manifest from %s: status code = %d", url, resp.StatusCode)
	}

	var app App
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return nil, ee.Wrapf(err, "cannot decode app manifest from %s", url)
	}

	app.Name = name

	return &app, nil
}

const KittyRegistryEnv = "KITTY_REGISTRY"
const DefaultKittyRegistry = "https://raw.githubusercontent.com/ImSingee/kitty-registry/master/apps/"

func GetRegistry() (string, error) {
	// from environment variable
	if r := os.Getenv(KittyRegistryEnv); r != "" {
		return r, nil
	}

	// from config file
	c, err := config.GetKittyConfig("")
	if err != nil && !ee.Is(err, os.ErrNotExist) {
		return "", ee.Wrap(err, "cannot load kitty config")
	}
	registryVal := c["registry"].Val()
	if registryVal != nil {
		if r, ok := registryVal.(string); !ok {
			return "", ee.New("invalid registry value (type is not string) in kitty config")
		} else if r != "" {
			return r, nil
		}
	}

	// default
	return DefaultKittyRegistry, nil
}

func (a *App) DownloadUnknownVersionTo(version string, dst string) error {
	if a.TryGoInstall != "" {
		return goInstallTo(version, a.TryGoInstall, dst, true)
	}

	return ee.Errorf("unknown app version %s", version)
}

func (v *Version) DownloadTo(dst string) error {
	return v.downloadTo(dst)
}

func (v *Version) downloadTo(dst string) error {
	osKey := GetCurrentBinKey()

	if u := v.Bin[osKey]; u != "" {
		return downloadFileTo(u, dst, 0755, true)
	}

	if v.GoInstall != "" {
		return goInstallTo(v.Version, v.GoInstall, dst, true)
	}

	return ee.Errorf("cannot find app version = %s on os %s", v.Version, osKey)
}

func downloadFileTo(url string, dst string, perm os.FileMode, showProgress bool) error {
	if showProgress {
		pp.Println("Download", url, "...")
	}

	_, err := mkdirFor(dst)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return ee.Wrapf(err, "cannot download file from %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ee.Errorf("cannot download file from %s: status code = %d", url, resp.StatusCode)
	}

	_ = os.Remove(dst)

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return ee.Wrapf(err, "cannot create file %s", dst)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return ee.Wrapf(err, "cannot write data to %s", dst)
	}

	err = f.Close()
	if err != nil {
		return ee.Wrapf(err, "cannot save and close file %s", dst)
	}

	return nil
}

func goInstallTo(version string, pkg string, dst string, showProgress bool) error {
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
