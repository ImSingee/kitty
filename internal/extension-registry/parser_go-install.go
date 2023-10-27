package extregistry

// GoInstallOptions is the options for go install
//
// - url: the url of the install path, e.g. "github.com/ImSingee/kitty/cmd/kitty"
type GoInstallOptions AnyOptions

func parseGoInstallOptions(in any) GoInstallOptions {
	return asAnyOptionsOrKey(in, "url")
}

func (o GoInstallOptions) ToDownloader() Installer {
	return nil
}
