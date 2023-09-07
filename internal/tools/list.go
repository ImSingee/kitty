package tools

import (
	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/kitty/internal/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type listOptions struct{}

type tool struct {
	name    string
	version string
}

// returned map is [name: version] in normal case
// but version can be any string in fact
// and version can be a '-' to indicate it's an external tool
func (o *listOptions) getCurrentToolsMap() (map[string]string, error) {
	c, err := config.GetKittyConfig("")
	if err != nil {
		if ee.Is(err, os.ErrNotExist) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	tools_ := c["tools"].Val()
	if tools_ == nil {
		return nil, nil
	}
	tools, ok := tools_.(map[string]any)
	if !ok {
		return nil, ee.Errorf("invalid tools config: must be a string-to-string map")
	}

	result := make(map[string]string, len(tools))
	for name, version := range tools {
		version, ok := version.(string)
		if !ok {
			return nil, ee.Errorf("invalid tools config: must be a string-to-string map")
		}

		result[name] = version
	}

	return result, nil
}

// returned map is [name: version] in normal case
// but version can be '?' if it's not a symlink or isn't a valid kitty symlink
func (o *listOptions) getInstalledTools() (map[string]string, error) {
	w, err := os.Getwd()
	if err != nil {
		return nil, ee.Wrap(err, "cannot get working directory")
	}

	files, err := os.ReadDir(filepath.Join(w, ".kitty", ".bin"))
	if err != nil {
		return nil, ee.Wrap(err, "cannot read .kitty/.bin dir")
	}

	result := make(map[string]string, len(files))

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()

		if strings.HasPrefix(name, ".") || name == "kitty" {
			continue
		}

		link, _ := os.Readlink(filepath.Join(w, ".kitty", ".bin", name))
		link = filepath.Base(link)
		appName, appVersion, _ := strings.Cut(link, "@")

		if appName == name && appVersion != "" {
			// valid kitty symlink
			result[appName] = appVersion
		} else {
			result[appName] = "?"
		}
	}

	return result, nil
}

func convertToolsMapToToolsSlice(m map[string]string) []*tool {
	result := make([]*tool, 0, len(m))
	for name, version := range m {
		result = append(result, &tool{
			name:    name,
			version: version,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].name < result[j].name
	})

	return result
}

func cloneToolsMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for name, version := range m {
		result[name] = version
	}

	return result
}
