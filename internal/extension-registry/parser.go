package extregistry

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ImSingee/go-ex/ee"

	eroptions "github.com/ImSingee/kitty/internal/extension-registry/options"
)

func parseApp(name string, in eroptions.AnyOptions) (*App, error) {
	slog.Debug("parseApp", "in", in)

	a := &App{
		Name: name,
	}

	if len(in) == 0 {
		return a, ee.New("invalid or malformed app manifest")
	}

	// parse tags
	if tags := in["tags"].Val(); tags != nil {
		tags, ok := tags.(map[string]any)
		if !ok {
			return nil, ee.Errorf("invalid `tags` key in app manifest")
		}

		parsedTags := make(map[string]string, len(tags))
		for tag, v := range tags {
			v, ok := v.(string)
			if !ok {
				parsedTags[tag] = "x invalid tag value"
			} else {
				parsedTags[tag] = v
			}
		}

		a.Tags = parsedTags
	}

	// parse versions
	if versions := in["versions"].Val(); versions != nil {
		versions, ok := versions.(map[string]any)
		if !ok {
			return nil, ee.Errorf("invalid `versions` key in app manifest")
		}

		parsedVersions := make(map[string]eroptions.AnyOptions, len(versions))

		for version, v := range versions {
			slog.Debug("parseVersion", "in", v)

			v, ok := v.(map[string]any)
			if !ok {
				parsedVersions[version] = eroptions.AsAnyOptions(map[string]any{
					"error": "invalid version value",
				})
			} else {
				parsedVersions[version] = eroptions.AsAnyOptions(v)
			}
		}

		a.Versions = parsedVersions
	}

	installOptions := eroptions.AnyOptions{}
	for k, v := range in {
		if strings.HasPrefix(k, "try-") {
			installOptions[k[4:]] = v
		}
	}
	a.InstallOptions = installOptions

	slog.Debug("parseApp", "parsed", a)

	return a, nil
}

var ErrVersionNotExist = fmt.Errorf("version not exist")

func (a *App) parseVersion(version string) (*Version, error) {
	vIn, ok := a.Versions[version]
	if !ok {
		return nil, ee.Wrapf(ErrVersionNotExist, "unknown version %s", version)
	}

	if err := eroptions.CastToError(vIn); err != nil {
		return nil, ee.Errorf("version is invalid or unsupported in current kitty version")
	}

	return &Version{
		App:            a,
		Version:        version,
		InstallOptions: vIn,
	}, nil
}
