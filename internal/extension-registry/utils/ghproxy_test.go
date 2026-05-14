package erutils

import "testing"

func TestApplyGitHubProxy(t *testing.T) {
	t.Run("leaves url unchanged when GHPROXY is not set", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "")

		got, err := ApplyGitHubProxy("https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz")
		if err != nil {
			t.Fatalf("ApplyGitHubProxy returned error: %v", err)
		}

		want := "https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz"
		if got != want {
			t.Fatalf("unexpected url:\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("prefixes github release urls", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "https://ghfast.top")

		got, err := ApplyGitHubProxy("https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz")
		if err != nil {
			t.Fatalf("ApplyGitHubProxy returned error: %v", err)
		}

		want := "https://ghfast.top/https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz"
		if got != want {
			t.Fatalf("unexpected url:\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("prefixes raw githubusercontent urls", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "https://ghfast.top")

		got, err := ApplyGitHubProxy("https://raw.githubusercontent.com/ImSingee/kitty-registry/master/apps/kitty/manifest.json")
		if err != nil {
			t.Fatalf("ApplyGitHubProxy returned error: %v", err)
		}

		want := "https://ghfast.top/https://raw.githubusercontent.com/ImSingee/kitty-registry/master/apps/kitty/manifest.json"
		if got != want {
			t.Fatalf("unexpected url:\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("prefixes gist hosts", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "https://ghfast.top")

		tests := map[string]string{
			"https://gist.github.com/user/hash":                "https://ghfast.top/https://gist.github.com/user/hash",
			"https://gist.githubusercontent.com/user/hash/raw": "https://ghfast.top/https://gist.githubusercontent.com/user/hash/raw",
		}
		for in, want := range tests {
			got, err := ApplyGitHubProxy(in)
			if err != nil {
				t.Fatalf("ApplyGitHubProxy(%q) returned error: %v", in, err)
			}
			if got != want {
				t.Fatalf("unexpected url for %q:\nwant: %s\ngot:  %s", in, want, got)
			}
		}
	})

	t.Run("leaves non github urls unchanged", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "https://ghfast.top")

		got, err := ApplyGitHubProxy("https://example.com/archive.tar.gz")
		if err != nil {
			t.Fatalf("ApplyGitHubProxy returned error: %v", err)
		}

		want := "https://example.com/archive.tar.gz"
		if got != want {
			t.Fatalf("unexpected url:\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("trims trailing slash from proxy", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, " https://ghfast.top/ ")

		got, err := ApplyGitHubProxy("https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz")
		if err != nil {
			t.Fatalf("ApplyGitHubProxy returned error: %v", err)
		}

		want := "https://ghfast.top/https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz"
		if got != want {
			t.Fatalf("unexpected url:\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("rejects invalid proxy", func(t *testing.T) {
		t.Setenv(GitHubProxyEnv, "ghfast.top")

		if _, err := ApplyGitHubProxy("https://github.com/ImSingee/kitty/releases/download/v1/kitty.tar.gz"); err == nil {
			t.Fatal("expected error for invalid GHPROXY")
		}
	})
}
