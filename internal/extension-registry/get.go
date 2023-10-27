package extregistry

import (
	"net/http"
	"net/url"
	"os"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ysmood/gson"

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

	v, err := a.parseVersion(version)
	if err != nil {
		return a, nil, ee.Wrapf(err, "cannot parse version %s of app %s", version, app)
	}

	return a, v, nil
}

func GetApp(app string) (*App, error) {
	r, err := GetRegistry()
	if err != nil {
		return nil, ee.Wrap(err, "cannot get registry")
	}

	urlForApp, err := url.JoinPath(r, "apps", app, "manifest.json")
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

	app, err := parseApp(name, gson.New(resp.Body).Map())
	if err != nil {
		return nil, ee.Wrapf(err, "cannot parse app manifest from %s", url)
	}

	return app, nil
}

const KittyRegistryEnv = "KITTY_REGISTRY"
const DefaultKittyRegistry = "https://raw.githubusercontent.com/ImSingee/kitty-registry/master/"

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
