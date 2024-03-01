package erutils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Unzip src (zip file) to dest (dir)
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	_ = os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		err := unzipFile(f, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

// unzipFile 处理ZIP文件中的每个文件或目录
func unzipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)

	// create dir
	if f.FileInfo().IsDir() {
		_ = os.MkdirAll(path, f.Mode())
		return nil
	}

	// create file
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, rc)
	return err
}
