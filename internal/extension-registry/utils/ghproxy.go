package erutils

import (
	"net/url"
	"os"
	"strings"

	"github.com/ImSingee/go-ex/ee"
)

const GitHubProxyEnv = "GHPROXY"

var githubProxyHosts = map[string]struct{}{
	"github.com":                 {},
	"gist.github.com":            {},
	"gist.githubusercontent.com": {},
	"raw.githubusercontent.com":  {},
}

func ApplyGitHubProxy(rawURL string) (string, error) {
	proxy := strings.TrimSpace(os.Getenv(GitHubProxyEnv))
	if proxy == "" {
		return rawURL, nil
	}

	proxy = strings.TrimRight(proxy, "/")
	parsedProxy, err := url.Parse(proxy)
	if err != nil || parsedProxy.Scheme == "" || parsedProxy.Host == "" || !isHTTPURL(parsedProxy) {
		return "", ee.Errorf("invalid %s value %q", GitHubProxyEnv, os.Getenv(GitHubProxyEnv))
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Host == "" {
		return "", ee.Errorf("invalid url %q", rawURL)
	}
	if !isHTTPURL(parsedURL) {
		return rawURL, nil
	}

	if _, ok := githubProxyHosts[strings.ToLower(parsedURL.Hostname())]; !ok {
		return rawURL, nil
	}

	return proxy + "/" + rawURL, nil
}

func isHTTPURL(u *url.URL) bool {
	return u.Scheme == "http" || u.Scheme == "https"
}
