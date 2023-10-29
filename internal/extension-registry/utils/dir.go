package erutils

import (
	"os"
	"path/filepath"

	"github.com/ImSingee/go-ex/ee"
)

func MkdirFor(file string) (string, error) {
	d, err := filepath.Abs(file)
	if err != nil {
		return "", ee.Wrapf(err, "cannot get absolute path of %s", file)
	}

	d = filepath.Dir(d)

	err = os.MkdirAll(d, 0755)
	if err != nil {
		return "", ee.Wrapf(err, "cannot create directory %s", d)
	}

	return d, nil
}
