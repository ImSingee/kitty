package extregistry

import "github.com/ImSingee/go-ex/ee"

func parseApp(name string, in AnyOptions) (*App, error) {
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

		parsedVersions := make(map[string]AnyOptions, len(versions))

		for version, v := range versions {
			v, ok := v.(map[string]any)
			if !ok {
				parsedVersions[version] = asAnyOptions(map[string]any{
					"error": "invalid version value",
				})
			} else {
				parsedVersions[version] = asAnyOptions(v)
			}
		}
	}

	// parse try-go-install options
	if tryGoInstall := in["try-go-install"].Val(); tryGoInstall != nil {
		a.TryGoInstall = parseGoInstallOptions(tryGoInstall)
	}

	return a, nil
}

func (a *App) parseVersion(version string) (*Version, error) {
	vIn, ok := a.Versions[version]
	if !ok {
		return nil, ee.Errorf("unknown version %s", version)
	}
	if err := vIn["error"].Val(); err != nil {
		return nil, ee.Errorf("version is invalid or unsupported in current kitty version")
	}

	return &Version{
		App:       a,
		Version:   version,
		Bin:       parseBinOptions(vIn["bin"]),
		GoInstall: parseGoInstallOptions(vIn["go-install"]),
	}, nil
}
