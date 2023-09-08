package lintstaged

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/go-ex/exstrings"
	"github.com/ImSingee/go-ex/mr"
	"github.com/ImSingee/go-ex/set"
	"github.com/ysmood/gson"

	"github.com/ImSingee/kitty/internal/config"
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
	Command string // command show to user

	Absolute bool
	NoArgs   bool
	Prepend  string // prepend to each file

	execCommand string // real command to execute
}

func searchConfigs(cwd, gitDir, configPath string) ([]*Config, error) {
	slog.Debug("Searching for configuration files...")

	// Use only explicit config path instead of discovering multiple
	if configPath != "" {
		slog.Debug("Using single configuration path...", "path", configPath)

		config, err := loadConfig(configPath)
		if err != nil {
			return nil, err
		}
		if config == nil {
			return nil, errors.New("cannot load configuration")
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

var validConfigNames = append(config.ConfigFileNames,
	".lintstagedrc",
	".lintstagedrc.json",
	// TODO yaml config files
	// TODO js config files
)

// loadConfig read and parse config file
//
// it may return (nil, nil) when the config can be ignored
func loadConfig(file string) (*Config, error) {
	in, err := configLoad(file)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
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

// configLoad loads config file and convert it to gson.JSON
//
// the returned map may be nil if it contains no config and no parse error
// and if it's not nil, it always contains a "files" key with type map[string]any
func configLoad(filename string) (map[string]gson.JSON, error) {
	name := filepath.Base(filename)
	if exstrings.InStringList(config.ConfigFileNames, name) { // kitty config
		return kittyConfigLoad(filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, ee.Wrap(err, "cannot read config file")
	}

	if strings.HasSuffix(name, "rc") || strings.HasSuffix(name, ".json") {
		return jsonLoad(content)
	}

	// // TODO support yaml
	//	// TODO support js

	return nil, ee.Errorf("unsupported config filename: %s", name)
}

// kittyConfigLoad loads kitty config file and extract the `lint-staged` key to map[string]gson.JSON
//
// the returned map may be nil if it doesn't contain the `lint-staged` key
// and if it's not nil, the returned map always contains a "files" key with type map[string]any
func kittyConfigLoad(filename string) (map[string]gson.JSON, error) {
	kc, err := config.ReadKittyConfig(filename)
	if err != nil {
		return nil, ee.Wrap(err, "cannot read kitty config file")
	}

	ls, ok := kc["lint-staged"]
	if !ok {
		return nil, nil
	}

	return loadConfigJSON(ls.Map())
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

	return loadConfigJSON(gson.New(in).Map())
}

func loadConfigJSON(m map[string]gson.JSON) (map[string]gson.JSON, error) {
	if len(m) == 0 {
		return nil, ee.New("empty config")
	}

	_, containsValidFilesKey := m["files"].Val().(map[string]any)
	if !containsValidFilesKey {
		m = map[string]gson.JSON{
			"files": gson.New(m),
		}
	}

	return m, nil
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

	if strings.HasPrefix(cmd, "[") { // custom options
	loop:
		for {
			switch {
			case strings.HasPrefix(cmd, "[absolute]"):
				result.Absolute = true
				cmd = strings.TrimPrefix(cmd, "[absolute]")
			case strings.HasPrefix(cmd, "[noArgs]"):
				result.NoArgs = true
				cmd = strings.TrimPrefix(cmd, "[noArgs]")
			case strings.HasPrefix(cmd, "[prepend ") || strings.HasPrefix(cmd, "[prepend="):
				// find next ]
				i := strings.Index(cmd, "]")
				if i < 0 {
					return nil, fmt.Errorf("command `%s` contains invalid prepend option", cmdIn)
				}

				prepend := strings.TrimSpace(cmd[len("[prepend "):i])
				cmd = cmd[i+1:]

				result.Prepend = prepend
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
	}

	cmd = strings.TrimSpace(cmd)

	if strings.HasPrefix(cmd, "@") {
		cmd = "kitty " + cmd
	}

	if strings.HasPrefix(cmd, "kitty ") {
		// change to self's absolute path (to avoid conflict)
		kitty, err := os.Executable()
		if err != nil {
			return nil, ee.Wrap(err, "cannot get current kitty's executable path")
		}
		result.execCommand = kitty + cmd[5:]
	} else {
		// TODO use .kitty/.bin/xxx as first priority
	}

	result.Command = cmd

	if result.execCommand == "" {
		result.execCommand = result.Command
	}

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
