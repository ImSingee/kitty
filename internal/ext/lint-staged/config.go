package lintstaged

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exstrings"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/set"
	"github.com/ysmood/gson"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	Path  string
	Rules []*Rule
}

type Rule struct {
	Glob     string
	Commands []*Command
}

type Command struct {
	Command  string
	Absolute bool
	NoArgs   bool
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

	in, err := jsonLoad(content)
	if err != nil {
		return nil, err
	}

	files := in["files"].Map()

	config := &Config{
		Path:  file,
		Rules: make([]*Rule, 0, len(files)),
	}

	for k, v := range files {
		vv, err := parseConfigRuleEntry(k, v)
		if err != nil {
			return nil, err
		}
		config.Rules = append(config.Rules, vv)
	}

	sort.Slice(config.Rules, func(i, j int) bool {
		return strings.Compare(config.Rules[i].Glob, config.Rules[j].Glob) < 0
	})

	return config, err

}

// jsonLoad loads json file and convert it to map[string]gson.JSON
//
// the returned map always contains a "files" key with type map[string]any
func jsonLoad(data []byte) (map[string]gson.JSON, error) {
	in := map[string]any{}

	err := json.Unmarshal(data, &in)
	if err != nil {
		return nil, err
	}

	_, containsValidFilesKey := in["files"].(map[string]any)
	if !containsValidFilesKey {
		in = map[string]any{
			"files": in,
		}
	}

	return gson.New(in).Map(), nil
}

func numberOfLevels(p string) int {
	return strings.Count(p, string(filepath.Separator))
}

func parseConfigRuleEntry(key string, v gson.JSON) (*Rule, error) {
	rule := &Rule{
		Glob:     key,
		Commands: nil,
	}

	switch vv := v.Val().(type) {
	case nil:
		return nil, fmt.Errorf("invalid nil command for %s", key)
	case string:
		cmd, err := parseStringCommand(vv)
		if err != nil {
			return nil, ee.Wrapf(err, "cannot parse command `%s`", vv)
		}
		rule.Commands = []*Command{cmd}
		return rule, nil
	case []any:
		result := make([]*Command, 0, len(vv))
		for i, vvv := range vv {
			s, err := parseRule(vvv)
			if err != nil {
				return nil, fmt.Errorf("invalid command for %s.%d: %w", key, i+1, err)
			}
			result = append(result, s)
		}
		rule.Commands = result
		return rule, nil
	default:
		return nil, fmt.Errorf("invalid value type (must be string or string list) for key %s", key)
	}
}

func parseRule(v any) (*Command, error) {
	// only support string now
	if str, ok := v.(string); ok {
		return parseStringCommand(str)
	}

	return nil, fmt.Errorf("invalid value type (must be string) for rule")
}

func parseStringCommand(cmd string) (*Command, error) {
	result := &Command{}

	cmdIn := cmd // backup for error message

	if !strings.HasPrefix(cmd, "[") { // no options
		return &Command{Command: cmd}, nil
	}

loop:
	for {
		switch {
		case strings.HasPrefix(cmd, "[absolute]"):
			result.Absolute = true
			cmd = strings.TrimPrefix(cmd, "[absolute]")
		case strings.HasPrefix(cmd, "[noArgs]"):
			result.NoArgs = true
			cmd = strings.TrimPrefix(cmd, "[noArgs]")
		default:
			break loop
		}
	}

	if strings.HasPrefix(cmd, "[") || strings.HasPrefix(cmd, " [") {
		return nil, fmt.Errorf("command `%s` contains unknown options", cmdIn)
	}

	if result.Absolute && result.NoArgs {
		return nil, fmt.Errorf("command `%s` cannot have both [absolute] and [noArgs] options", cmdIn)
	}

	result.Command = strings.TrimSpace(cmd)

	return result, nil
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
