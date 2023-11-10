package erutils

import (
	"os"
	"syscall"

	"github.com/ImSingee/go-ex/ee"
)

func Rename(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err == nil {
		return nil
	}

	// os.Rename guarantees it returns an *os.LinkError
	if err.(*os.LinkError).Err != syscall.EXDEV { //nolint:forcetypeassert
		return err
	}

	// fallback to copy and remove
	return moveFile(oldpath, newpath)
}

func moveFile(oldpath, newpath string) error {
	err := copyFile(oldpath, newpath, true)
	if err != nil {
		return ee.Wrapf(err, "rename %s -> %s: cannot copy file", oldpath, newpath)
	}
	err = os.Remove(oldpath)
	if err != nil {
		return ee.Wrapf(err, "rename %s -> %s: cannot remove old file", oldpath, newpath)
	}

	return nil
}
