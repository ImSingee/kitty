package config

import (
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exjson"
	"github.com/ysmood/gson"
	"os"
	"path/filepath"
)

var ConfigFileNames = []string{
	".kittyrc",
	".kittyrc.json",
	"kitty.config.json",
}

func ReadKittyConfig(filename string) (map[string]gson.JSON, error) {
	// only parse json now

	var obj map[string]any
	err := exjson.Read(filename, &obj)
	if err != nil {
		return nil, err
	}

	return gson.New(obj).Map(), nil
}

// GetKittyConfig will find the kitty config file in the given directory
//
// if dir is empty, it will use the current working directory
// if no config file found, it will return an error wrapped os.ErrNotExist, use ee.Is to check
func GetKittyConfig(dir string) (map[string]gson.JSON, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("cannot get working directory: %w", err)
		}
	}

	for _, name := range ConfigFileNames {
		filename := filepath.Join(dir, name)
		c, err := ReadKittyConfig(filename)
		if err != nil {
			if ee.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("cannot read kitty config file %s: %w", filename, err)
		}

		return c, nil
	}

	return nil, ee.Wrap(os.ErrNotExist, "cannot find kitty config file")
}
