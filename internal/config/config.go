package config

import (
	"encoding/json"
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
// if no config file found, it will return an error wrapped ErrNotExist, use IsNotExist to check
func GetKittyConfig(dir string) (map[string]gson.JSON, error) {
	filename, c, err := getKittyConfig(dir)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return nil, ee.Wrap(ErrNotExist, "cannot find kitty config file")
	}

	return c, nil
}

// TODO report error if there's more than one config

func getKittyConfig(dir string) (string, map[string]gson.JSON, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", nil, fmt.Errorf("cannot get working directory: %w", err)
		}
	}

	for _, name := range ConfigFileNames {
		filename := filepath.Join(dir, name)
		c, err := ReadKittyConfig(filename)
		if err != nil {
			if ee.Is(err, os.ErrNotExist) {
				continue
			}

			return filename, nil, fmt.Errorf("cannot read kitty config file %s: %w", filename, err)
		}

		return filename, c, nil
	}

	return "", nil, nil
}

func PatchKittyConfig(dir string, patch func(map[string]gson.JSON) error) error {
	filename, c, err := getKittyConfig(dir)
	if err != nil {
		return err
	}

	if filename == "" { // config not exist, generate one
		filename = filepath.Join(dir, ConfigFileNames[0])
		c = make(map[string]gson.JSON)
	}

	err = patch(c)
	if err != nil {
		return err
	}

	err = saveKittyConfig(filename, c)
	if err != nil {
		return err
	}

	return nil
}

func saveKittyConfig(filename string, c map[string]gson.JSON) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return ee.Wrap(err, "cannot json encode config")
	}

	return os.WriteFile(filename, data, 0644)
}
