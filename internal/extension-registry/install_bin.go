package extregistry

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/pp"
)

type distInstaller struct {
	o BinOptions
}

func (d *distInstaller) Install(o *InstallOptions) error {
	if err := d.o["error"].Val(); err != nil {
		return fmt.Errorf("%v", err)
	}

	osKey := GetCurrentBinKey()
	url := d.o[osKey].Val()

	url := d.o["url"].Val()
	if url == nil {
		return ErrInstallerNotApplicable
	}

	urlString, ok := url.(string)
	if !ok {
		return ee.New("`url` key in bin options is not string")
	}

	return downloadFileTo(urlString, o.To, 0755, o.ShowProgress)
}

func downloadFileTo(url string, dst string, perm os.FileMode, showProgress bool) error {
	if showProgress {
		pp.Println("Download", url, "...")
	}

	_, err := mkdirFor(dst)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return ee.Wrapf(err, "cannot download file from %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ee.Errorf("cannot download file from %s: status code = %d", url, resp.StatusCode)
	}

	_ = os.Remove(dst)

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return ee.Wrapf(err, "cannot create file %s", dst)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return ee.Wrapf(err, "cannot write data to %s", dst)
	}

	err = f.Close()
	if err != nil {
		return ee.Wrapf(err, "cannot save and close file %s", dst)
	}

	return nil
}
