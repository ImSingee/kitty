package extregistry

type App struct {
	Name     string                `json:"-"`
	Tags     map[string]string     `json:"tags"`     // key is tag name, e.g. "latest", value is version string, e.g. "1.0.0"
	Versions map[string]AnyOptions `json:"versions"` // key is version string, e.g. "1.0.0"

	TryGoInstall GoInstallOptions `json:"try-go-install"` // use go build to install for unknown version
}

type Version struct {
	App     *App   `json:"-"`
	Version string `json:"-"`

	Bin       BinOptions       `json:"bin"`        // directly bin download
	GoInstall GoInstallOptions `json:"go-install"` // use go build to install
}
