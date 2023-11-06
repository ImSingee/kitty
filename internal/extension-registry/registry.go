package extregistry

import eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"

type App struct {
	Name     string                          `json:"-"`
	Tags     map[string]string               `json:"tags"`     // key is tag name, e.g. "latest", value is version string, e.g. "1.0.0"
	Versions map[string]eroptions.AnyOptions `json:"versions"` // key is version string, e.g. "1.0.0"

	InstallOptions eroptions.AnyOptions
}

type Version struct {
	App     *App   `json:"-"`
	Version string `json:"-"`

	InstallOptions eroptions.AnyOptions
}
