package tmpl

import (
	"runtime"
	"strings"
	"text/template"

	"github.com/ImSingee/go-ex/ee"
	semver "github.com/Masterminds/semver/v3"

	"github.com/ImSingee/kitty/internal/extension-registry/installer"
)

func Render(tmpl string, options *installer.InstallOptions) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", ee.Wrap(err, "invalid template")
	}

	w := strings.Builder{}
	err = t.Execute(&w, toTemplateOptions(options))
	if err != nil {
		return "", ee.Wrap(err, "cannot render template")
	}
	return w.String(), nil
}

func toTemplateOptions(options *installer.InstallOptions) map[string]interface{} {
	v := ""
	if sv, _ := semver.NewVersion(options.Version); sv != nil {
		v = sv.String()
	}

	return map[string]interface{}{
		"version": options.Version,
		"semver":  v,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
}
