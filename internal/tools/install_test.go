package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ImSingee/kitty/internal/extension-registry/binkey"
)

func TestEnsureInstalledUpdatesOutdatedConfiguredTool(t *testing.T) {
	root := t.TempDir()
	binKey := string(binkey.GetCurrentBinKey())

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/tool/manifest.json":
			_, _ = fmt.Fprintf(w, `{
				"tags": {"latest": "2.0.0"},
				"versions": {
					"2.0.0": {
						"bin": {
							"%s": {"url": "%s/tool-bin"}
						}
					}
				}
			}`, binKey, server.URL)
		case "/tool-bin":
			_, _ = w.Write([]byte("#!/bin/sh\necho new\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	configJSON := fmt.Sprintf(`{"registry":%q,"tools":{"tool":"latest"}}`, server.URL+"/")
	err := os.WriteFile(filepath.Join(root, ".kittyrc.json"), []byte(configJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(root, ".kitty", ".bin")
	err = os.MkdirAll(filepath.Join(binDir, "."+binKey), 0755)
	if err != nil {
		t.Fatal(err)
	}

	oldTarget := filepath.Join("."+binKey, "tool@1.0.0")
	err = os.WriteFile(filepath.Join(binDir, oldTarget), []byte("#!/bin/sh\necho old\n"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(oldTarget, filepath.Join(binDir, "tool"))
	if err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(root, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	previousWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWd)
	})

	if err := EnsureInstalled(root, "tool"); err != nil {
		t.Fatal(err)
	}

	wantTarget := filepath.Join("."+binKey, "tool@2.0.0")
	gotTarget, err := os.Readlink(filepath.Join(binDir, "tool"))
	if err != nil {
		t.Fatal(err)
	}
	if gotTarget != wantTarget {
		t.Fatalf("tool symlink = %q, want %q", gotTarget, wantTarget)
	}

	data, err := os.ReadFile(filepath.Join(binDir, wantTarget))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "echo new") {
		t.Fatalf("installed binary content = %q, want new binary", string(data))
	}

	configData, err := os.ReadFile(filepath.Join(root, ".kittyrc.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsedConfig struct {
		Tools map[string]string `json:"tools"`
	}
	if err := json.Unmarshal(configData, &parsedConfig); err != nil {
		t.Fatal(err)
	}
	if parsedConfig.Tools["tool"] != "2.0.0" {
		t.Fatalf("config tool version = %q, want 2.0.0", parsedConfig.Tools["tool"])
	}
}

func TestEnsureInstalledIgnoresUnconfiguredRequestedTool(t *testing.T) {
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, ".kittyrc.json"), []byte(`{"tools":{}}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstalled(root, "missing"); err != nil {
		t.Fatal(err)
	}
}
