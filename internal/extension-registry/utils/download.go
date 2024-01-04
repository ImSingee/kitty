package erutils

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
)

func downloadFileTo(url string, w io.Writer, showProgress bool) error {
	if showProgress { // TODO progress bar
		pp.Println("Download", url, "...")
	}

	resp, err := http.Get(url)
	if err != nil {
		return ee.Wrapf(err, "cannot download file from %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ee.Errorf("cannot download file from %s: status code = %d", url, resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return ee.Wrap(err, "cannot write data")
	}

	return nil
}

func DownloadFileTo(url string, dst string, perm os.FileMode, showProgress bool) error {
	_, err := MkdirFor(dst)
	if err != nil {
		return err
	}

	_ = os.Remove(dst)
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return ee.Wrapf(err, "cannot create file %s", dst)
	}
	defer f.Close()

	err = downloadFileTo(url, f, showProgress)
	if err != nil {
		return ee.Wrapf(err, "cannot download file to %s", dst)
	}

	err = f.Close()
	if err != nil {
		return ee.Wrapf(err, "cannot save and close file %s", dst)
	}

	return nil
}

func DownloadFileToTemp(url string, temppattern string, showProgress bool) (string, error) {
	f, err := os.CreateTemp("", temppattern)
	if err != nil {
		return "", ee.Wrap(err, "cannot create temp file")
	}
	defer f.Close()

	slog.Debug("DownloadFileToTemp", "url", url, "to", f.Name())
	err = downloadFileTo(url, f, showProgress)
	if err != nil {
		return "", ee.Wrap(err, "cannot download file to [tempfile]")
	}

	err = f.Close()
	if err != nil {
		return "", ee.Wrap(err, "cannot save and close file [tempfile]")
	}

	return f.Name(), nil
}
