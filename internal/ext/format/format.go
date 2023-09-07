package format

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/ImSingee/go-ex/ee"
)

type options struct {
	files []string

	allowUnknown bool
}

func (o *options) run() error {
	errs := make([]error, 0, 6)

	for _, file := range o.files {
		err := o.format(file)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return ee.Errorf("some error during format: %v", errs)
	}

	return nil
}

func (o *options) format(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	name := filepath.Base(file)
	ext := filepath.Ext(name)

	switch ext {
	case ".json":
		return o.jsonFormat(file, content)
	}

	if !o.allowUnknown {
		return ee.Errorf("Cannot format unknown file %s", file)
	}

	return nil
}

type Formatter interface {
	Format(filename string, content []byte) ([]byte, error)
}

func (o *options) doFormat(filename string, content []byte, formatter Formatter) error {
	if content == nil {
		c, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		content = c
	}

	newContent, err := formatter.Format(filename, content)
	if err != nil {
		return err
	}

	changed := !bytes.Equal(content, newContent)

	if changed {
		err := os.WriteFile(filename, newContent, 0)
		if err != nil {
			return ee.Wrapf(err, "cannot overwrite file %s", filename)
		}
	}

	return nil
}
