package main

import "fmt"

var (
	version = "DEV"
	commit  = ""
	buildAt = ""
)

func getVersionString() string {
	return fmt.Sprintf("%s\nCommit: %s\nBuild At: %s", version, commit, buildAt)
}
