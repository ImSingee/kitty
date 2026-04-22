package kittyversion

import (
	semver "github.com/Masterminds/semver/v3"
	"github.com/ysmood/gson"

	"github.com/ImSingee/kitty/internal/version"
)

func ParseRequired(c map[string]gson.JSON) string {
	if c == nil {
		return ""
	}

	kitty, kittyExists := c["kitty"]
	if !kittyExists {
		return ""
	}

	switch kittyVal := kitty.Val().(type) {
	case string:
		return kittyVal
	case map[string]any:
		if v, ok := kittyVal["version"].(string); ok {
			return v
		}

		return ""
	default:
		return ""
	}
}

func CurrentSatisfies(requiredVersion string) (bool, error) {
	return Satisfies(version.Version(), requiredVersion)
}

func Satisfies(currentVersion string, requiredVersion string) (bool, error) {
	if currentVersion == version.DevVersion {
		return true, nil
	}

	constraints, err := semver.NewConstraint(requiredVersion)
	if err != nil {
		return false, err
	}

	sv, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, err
	}

	return constraints.Check(sv), nil
}
