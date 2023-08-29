package git

import (
	"strings"
)

type FileStatus struct {
	Name              string
	IndexStatus       byte
	WorkingTreeStatus byte
	RenameFrom        string // only exists for rename
}

func Status(dir string, allowSubmodule bool) ([]FileStatus, error) {
	args := []string{"status", "--porcelain"}
	if !allowSubmodule {
		args = append([]string{"-c", "submodule.recurse=false"}, args...)
	}

	result := R(dir, args)
	output := result.Output
	err := result.Err()

	if err != nil {
		return nil, err
	}

	return parseStatus(string(output))
}

func parseStatus(status string) ([]FileStatus, error) {
	lines := strings.Split(strings.TrimSpace(status), "\n")
	result := make([]FileStatus, 0, len(lines))

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		name := line[3:]

		renameFrom := ""
		left, right, isRename := strings.Cut(name, " -> ")
		if isRename {
			name = right
			renameFrom = left
		}

		result = append(result, FileStatus{
			Name:              name,
			IndexStatus:       line[0],
			WorkingTreeStatus: line[1],
			RenameFrom:        renameFrom,
		})
	}

	return result, nil
}
