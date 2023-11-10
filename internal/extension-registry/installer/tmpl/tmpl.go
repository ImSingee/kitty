package tmpl

import (
	"strings"
	"text/template"

	"github.com/ImSingee/go-ex/ee"
	"github.com/ImSingee/semver"

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
	}
}
