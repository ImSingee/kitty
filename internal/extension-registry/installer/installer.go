package installer

import (
	"fmt"

	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

type InstallOptions struct {
	Version      string
	To           string
	ShowProgress bool
}

var ErrInstallerNotApplicable = fmt.Errorf("installer is not available")

type Factory interface {
	// Key returns the key of the installer
	//
	// The key is used to identify the installer, and it's also
	// the key of the options in the extension registry.
	Key() string

	// GetInstaller returns the installer if it's available
	//
	// If the installer is not available, it will return (nil, some error)
	// (the returned error is only for inform purpose, it's safe to ignore)
	GetInstaller(o eroptions.AnyOptions) (Installer, error)
}

var ErrSkip = fmt.Errorf("skip")

type Installer interface {
	// Install will install the app to the given destination
	//
	// If the installer is not application, it will return ErrSkip
	Install(o *InstallOptions) error
}
