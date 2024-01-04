package version

import "fmt"

var (
	version = "DEV"
	commit  = ""
	buildAt = ""
)

func Version() string {
	return version
}
func Commit() string {
	return commit
}
func BuildAt() string {
	return buildAt
}

func GetVersionString() string {
	return fmt.Sprintf("%s\nCommit: %s\nBuild At: %s", version, commit, buildAt)
}
