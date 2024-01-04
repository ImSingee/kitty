package version

import (
	"fmt"

	"github.com/ImSingee/semver"
)

var (
	version = DevVersion
	commit  = ""
	buildAt = ""
)

const DevVersion = "DEV"

func Version() string {
	return version
}
func Commit() string {
	return commit
}
func BuildAt() string {
	return buildAt
}

func Semver() *semver.Version {
	v, err := semver.NewVersion(version)
	if err == nil {
		return v
	} else {
		return semver.NewVersionByParts(0, 0, 0)
	}
}

// LessThan 返回当前程序版本是否小于传入的版本
// 注：DEV 版本始终返回 false
func LessThan(v *semver.Version) bool {
	if version == "DEV" {
		return false
	}

	return Semver().LessThan(v)
}

func GetVersionString() string {
	return fmt.Sprintf("%s\nCommit: %s\nBuild At: %s", version, commit, buildAt)
}
