package extregistry

import (
	"github.com/ImSingee/semver"
)

func normalizeGoVersion(v string) string {
	if v == "" {
		return "latest"
	}

	if version, err := semver.NewVersion(v); err != nil {
		return "v" + version.String()
	}

	return v
}
