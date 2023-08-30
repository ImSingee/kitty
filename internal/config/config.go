package config

import (
	"github.com/ImSingee/go-ex/exjson"
	"github.com/ysmood/gson"
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
