package erutils

import (
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/ImSingee/go-ex/ee"
)

func copyFile(from, to string, force bool) error {
	from = filepath.Clean(from)
	to = filepath.Clean(to)
	if from == to {
		return nil
	}

	fromF, err := os.Open(from)
	if err != nil {
		return ee.Wrapf(err, "cannot open source file %s", from)
	}
	defer fromF.Close()

	stat, err := fromF.Stat()
	if err != nil {
		return ee.Wrapf(err, "cannot stat source file %s", from)
	}

	var toF *os.File
	if force { // 存在时覆盖
		toF, err = os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	} else { // 存在时报错
		toF, err = os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, stat.Mode())
	}

	if err != nil {
		return ee.Wrapf(err, "cannot open destination file %s", to)
	}
	defer toF.Close()

	_, err = io.Copy(toF, fromF)
	if err != nil {
		return ee.Wrap(err, "cannot copy data")
	}

	// change owner
	sysStat := stat.Sys().(*syscall.Stat_t) //nolint:forcetypeassert
	err = os.Chown(to, int(sysStat.Uid), int(sysStat.Gid))
	if err != nil {
		return ee.Wrap(err, "cannot change destination file owner")
	}

	return nil
}
