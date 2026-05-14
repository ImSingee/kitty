package extregistry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	erutils "github.com/ImSingee/kitty/internal/extension-registry/utils"
)

func TestGetAppFromUrlUsesGitHubProxy(t *testing.T) {
	const manifestURL = "https://raw.githubusercontent.com/ImSingee/kitty-registry/master/apps/tool/manifest.json"

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tags": {"latest": "1.0.0"},
			"versions": {
				"1.0.0": {"go-install": "example.com/tool@v1.0.0"}
			}
		}`))
	}))
	defer server.Close()

	t.Setenv(erutils.GitHubProxyEnv, server.URL)

	app, err := getAppFromUrl("tool", manifestURL)
	if err != nil {
		t.Fatalf("getAppFromUrl returned error: %v", err)
	}

	if app.Tags["latest"] != "1.0.0" {
		t.Fatalf("unexpected latest tag: %q", app.Tags["latest"])
	}

	wantPath := "/" + manifestURL
	if requestedPath != wantPath {
		t.Fatalf("unexpected proxied request path:\nwant: %s\ngot:  %s", wantPath, requestedPath)
	}
}
