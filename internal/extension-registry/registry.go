package extregistry

type App struct {
	Name     string              `json:"-"`
	Tags     map[string]string   `json:"tags"`     // key is tag name, e.g. "latest", value is version string, e.g. "1.0.0"
	Versions map[string]*Version `json:"versions"` // key is version string, e.g. "1.0.0"

	TryGoInstall string `json:"try-go-install"` // use go build to install for unknown version
}

type Version struct {
	Version string            `json:"-"`
	Bin     map[BinKey]string `json:"bin"` // directly bin download

	GoInstall string `json:"go-install"` // use go build to install
}
