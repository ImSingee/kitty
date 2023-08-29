package lintstaged

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exstrings"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/set"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	Path  string
	Rules map[string][]string
}

func searchConfigs(cwd, gitDir, configPath string, configObject *Config) ([]*Config, error) {
	slog.Debug("Searching for configuration files...")

	if configObject != nil {
		slog.Debug("Using single direct configuration object...")

		return []*Config{configObject}, nil
	}

	// Use only explicit config path instead of discovering multiple
	if configPath != "" {
		slog.Debug("Using single configuration path...", "path", configPath)

		config, err := loadConfig(configPath)
		if err != nil {
			return nil, err
		}
		config.Path = configPath

		return []*Config{config}, nil
	}

	cachedFiles, err := execGitZ([]string{"ls-files", "-z", "--full-name"}, gitDir)
	if err != nil {
		return nil, ee.Wrap(err, "cannot get list of known files")
	}

	//otherFiles, err := execGitZ([]string{"ls-files", "-z", "--full-name", "--others", "--exclude-standard"}, gitDir)
	//if err != nil {
	//	return nil, ee.Wrap(err, "cannot get list of uncommitted files")
	//}
	//possibleConfigFiles := mr.Flats(cachedFiles, otherFiles)

	possibleConfigFiles := cachedFiles
	possibleConfigFiles = mr.Filter(possibleConfigFiles, func(file string, _index int) bool {
		return exstrings.InStringList(validConfigNames, filepath.Base(file))
	})
	possibleConfigFiles = mr.Map(possibleConfigFiles, func(file string, _index int) string {
		return normalizePath(filepath.Join(gitDir, file))
	})
	possibleConfigFiles = mr.Filter(possibleConfigFiles, func(file string, _index int) bool {
		return strings.HasPrefix(file, cwd)
	})
	sort.Slice(possibleConfigFiles, func(i, j int) bool {
		return numberOfLevels(possibleConfigFiles[i]) > numberOfLevels(possibleConfigFiles[j])
	})

	slog.Debug("Found possible config files", "possibleConfigFiles", possibleConfigFiles, "possibleConfigFilesCount", len(possibleConfigFiles))

	configs := make([]*Config, 0, len(possibleConfigFiles))
	for _, file := range possibleConfigFiles {
		config, err := loadConfig(file)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, ee.Wrapf(err, "cannot load config file %s", file)
		}
		if config == nil {
			continue // ignore it
		}

		config.Path = file

		configs = append(configs, config)
	}

	if debug() {
		loadedConfigFiles := mr.Map(configs, func(config *Config, _index int) string {
			return config.Path
		})

		slog.Debug("Load config files", "loadedConfigFiles", loadedConfigFiles, "loadedConfigFilesCount", len(loadedConfigFiles))
	}

	configBasePaths := make(map[string]string, len(configs))
	for _, config := range configs {
		basePath := filepath.Dir(config.Path)
		if exist := configBasePaths[basePath]; exist != "" {
			return nil, fmt.Errorf("multiple config files found in the same directory: %s, files: [%s, %s]", basePath, configBasePaths[basePath], config.Path)
		}
		configBasePaths[basePath] = config.Path
	}

	return configs, nil
}

var validConfigNames = []string{
	//TODO KITTY_CONFIG_FILES
	".lintstagedrc",
	".lintstagedrc.json",
	".lintstagedrc.yaml",
	".lintstagedrc.yml",
	// TODO js config files
}

func loadConfig(file string) (*Config, error) {
	// TODO support yaml
	// TODO support js

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, ee.Wrap(err, "cannot read config file")
	}

	config := &Config{}

	in, err := jsonLoad(content)
	if err != nil {
		return nil, err
	}

	config.Path = file
	config.Rules = make(map[string][]string, len(in))
	for k, v := range in {
		vv, err := parseStringList(k, v)
		if err != nil {
			return nil, err
		}
		config.Rules[k] = vv
	}

	return config, err

}

func jsonLoad(data []byte) (map[string]any, error) {
	in := map[string]any{}

	err := json.Unmarshal(data, &in)

	return in, err
}

func numberOfLevels(p string) int {
	return strings.Count(p, string(filepath.Separator))
}

func parseStringList(key string, v any) ([]string, error) {
	switch vv := v.(type) {
	case nil:
		return nil, fmt.Errorf("invalid nil value for key %s", key)
	case string:
		return []string{vv}, nil
	case []any:
		result := make([]string, 0, len(vv))
		for _, vvv := range vv {
			s, ok := vvv.(string)
			if !ok {
				return nil, fmt.Errorf("invalid string list for key %s", key)
			}
			result = append(result, s)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("invalid value type (must be string or string list) for key %s", key)
	}
}

// groupFilesByConfig map file to specific (deepest level) config
func groupFilesByConfig(configs []*Config, files []string) map[*Config][]string {
	group := make(map[*Config][]string, len(configs))
	if len(configs) == 1 {
		group[configs[0]] = files
		return group
	}

	filesSet := set.New(files...)

	for _, config := range configs {
		d := filepath.Dir(config.Path) + string(filepath.Separator)

		filesSet.Do(func(s string) {
			if strings.HasPrefix(s, d) {
				group[config] = append(group[config], s)
				filesSet.Remove(s)
			}
		})
	}

	return group
}
