package git

import (
	"bytes"
	"os"
	"path/filepath"
)

func (g *G) Root() (string, error) {
	result := g.Run("rev-parse", "--show-toplevel")
	if err := result.Err(); err != nil {
		return "", err
	}

	root := string(bytes.TrimSpace(result.Output))

	return root, nil
}

func (g *G) IsRoot() (bool, error) {
	_, err := os.Stat(filepath.Join(g.Dir, ".git"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func GetRoot(dir string) (string, error) {
	return (&G{Dir: dir}).Root()
}

func IsRoot(dir string) (bool, error) {
	return (&G{Dir: dir}).IsRoot()
}
