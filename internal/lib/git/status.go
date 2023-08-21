package git

// Status returns the output of `git status --porcelain`
func Status() (string, error) {
	result := run("status", "--porcelain")

	return string(result.Output), result.Err()
}
