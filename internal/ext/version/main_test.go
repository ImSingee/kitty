package versionext

import (
	"strings"
	"testing"
)

func TestDescribeToVersion(t *testing.T) {
	tests := []struct {
		name     string
		describe string
		want     string
	}{
		{name: "semver tag", describe: "v1.2.3", want: "1.2.3"},
		{name: "semver describe", describe: "v1.2.3-4-gabcdef12-dirty", want: "1.2.3-4-gabcdef12-dirty"},
		{name: "plain commit", describe: "abcdef12-dirty", want: "0.0.0-abcdef12-dirty"},
		{name: "empty", describe: "", want: "0.0.0-unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeToVersion(tt.describe); got != tt.want {
				t.Fatalf("describeToVersion(%q) = %q, want %q", tt.describe, got, tt.want)
			}
		})
	}
}

func TestQuoteEnvValue(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "v1.2.3-4-gabcdef12", want: "v1.2.3-4-gabcdef12"},
		{value: "", want: "''"},
		{value: "go version go1.26.2 linux/amd64", want: "'go version go1.26.2 linux/amd64'"},
		{value: "can't", want: `'can'\''t'`},
	}

	for _, tt := range tests {
		if got := quoteEnvValue(tt.value); got != tt.want {
			t.Fatalf("quoteEnvValue(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestRenderEnv(t *testing.T) {
	env := renderEnv(Info{
		Version:        "1.2.3",
		GitCommit:      "abcdef1234567890",
		GitCommitShort: "abcdef12",
		GitTag:         "v1.2.3",
		GitDescribe:    "v1.2.3",
		GitDirty:       true,
		BuildTime:      "2026-05-20T06:00:00Z",
		BuildID:        "20260520060000-abcdef12",
		KittyVersion:   "DEV",
	})

	for _, line := range []string{
		"VERSION=1.2.3\n",
		"GITDIRTY=true\n",
		"KITTY_COMMIT=''\n",
	} {
		if !strings.Contains(env, line) {
			t.Fatalf("rendered env does not contain %q:\n%s", line, env)
		}
	}
}
