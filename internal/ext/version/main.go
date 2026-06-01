package versionext

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ImSingee/go-ex/ee"
	semver "github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"

	kittyversion "github.com/ImSingee/kitty/internal/version"
)

type options struct {
	output string
}

type Info struct {
	Version        string
	GitCommit      string
	GitCommitShort string
	GitTag         string
	GitDescribe    string
	GitDirty       bool
	BuildTime      string
	BuildID        string
	KittyVersion   string
	KittyCommit    string
	KittyBuildAt   string
}

const zeaburGitCommitSHAEnv = "ZEABUR_GIT_COMMIT_SHA"

func Commands() []*cobra.Command {
	o := &options{}

	cmd := &cobra.Command{
		Use:   "@version",
		Short: "print build version information as an env file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := o.run()
			if err != nil {
				return err
			}

			if o.output != "" {
				return os.WriteFile(o.output, []byte(content), 0644)
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), content)
			return err
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.output, "output", "o", "", "write env file to path instead of stdout")

	return []*cobra.Command{cmd}
}

func (o *options) run() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", ee.Wrap(err, "cannot get working directory")
	}

	info, err := Collect(wd, time.Now().UTC())
	if err != nil {
		return "", err
	}

	return renderEnv(info), nil
}

func Collect(dir string, now time.Time) (Info, error) {
	if info, ok := collectZeabur(now); ok {
		return info, nil
	}

	return collectGit(dir, now)
}

func collectZeabur(now time.Time) (Info, bool) {
	commit := strings.TrimSpace(os.Getenv(zeaburGitCommitSHAEnv))
	if commit == "" {
		return Info{}, false
	}

	commitShort := shortHash(commit)
	return completeInfo(Info{
		Version:        describeToVersion(commitShort),
		GitCommit:      commit,
		GitCommitShort: commitShort,
		GitDescribe:    commitShort,
		GitDirty:       false,
	}, now), true
}

func collectGit(dir string, now time.Time) (Info, error) {
	repo, err := openRepository(dir)
	if err != nil {
		return Info{}, err
	}

	head, err := repo.Head()
	if err != nil {
		return Info{}, ee.Wrap(err, "cannot get git HEAD")
	}

	commit := head.Hash().String()
	commitShort := shortHash(commit)

	gitTag, gitDescribe, err := describe(repo, head.Hash())
	if err != nil {
		return Info{}, err
	}

	dirty, err := isDirty(repo)
	if err != nil {
		return Info{}, err
	}
	if dirty {
		gitDescribe += "-dirty"
	}

	return completeInfo(Info{
		Version:        describeToVersion(gitDescribe),
		GitCommit:      commit,
		GitCommitShort: commitShort,
		GitTag:         gitTag,
		GitDescribe:    gitDescribe,
		GitDirty:       dirty,
	}, now), nil
}

func completeInfo(info Info, now time.Time) Info {
	info.BuildTime = now.Format(time.RFC3339)
	info.BuildID = now.Format("20060102150405") + "-" + info.GitCommitShort
	info.KittyVersion = kittyversion.Version()
	info.KittyCommit = kittyversion.Commit()
	info.KittyBuildAt = kittyversion.BuildAt()

	return info
}

func openRepository(dir string) (*git.Repository, error) {
	root, err := findGitRoot(dir)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpenWithOptions(root, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, ee.Wrapf(err, "cannot open git repository at %s", root)
	}

	return repo, nil
}

func findGitRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}

	dir, err := filepath.Abs(start)
	if err != nil {
		return "", ee.Wrapf(err, "cannot resolve path %s", start)
	}

	info, err := os.Stat(dir)
	if err != nil {
		return "", ee.Wrapf(err, "cannot stat path %s", dir)
	}
	if !info.IsDir() {
		dir = filepath.Dir(dir)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", ee.Wrapf(err, "cannot stat %s", filepath.Join(dir, ".git"))
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ee.Errorf("cannot find git repository from %s", start)
		}
		dir = parent
	}
}

func describe(repo *git.Repository, head plumbing.Hash) (string, string, error) {
	tagsByCommit, err := tagsByCommit(repo)
	if err != nil {
		return "", "", err
	}

	iter, err := repo.Log(&git.LogOptions{From: head})
	if err != nil {
		return "", "", ee.Wrap(err, "cannot read git log")
	}
	defer iter.Close()

	distance := 0
	headShort := head.String()[:8]
	var tagName string

	err = iter.ForEach(func(commit *object.Commit) error {
		if names := tagsByCommit[commit.Hash]; len(names) > 0 {
			tagName = names[0]
			return storer.ErrStop
		}

		distance++
		return nil
	})
	if err != nil && !errors.Is(err, storer.ErrStop) {
		return "", "", ee.Wrap(err, "cannot describe git HEAD")
	}

	if tagName == "" {
		return "", headShort, nil
	}

	if distance == 0 {
		return tagName, tagName, nil
	}

	return tagName, fmt.Sprintf("%s-%d-g%s", tagName, distance, headShort), nil
}

func tagsByCommit(repo *git.Repository) (map[plumbing.Hash][]string, error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, ee.Wrap(err, "cannot list git tags")
	}
	defer iter.Close()

	result := map[plumbing.Hash][]string{}
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		commitHash, ok := resolveTagCommit(repo, ref.Hash())
		if !ok {
			return nil
		}

		result[commitHash] = append(result[commitHash], ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, ee.Wrap(err, "cannot iterate git tags")
	}

	for commitHash := range result {
		sort.Strings(result[commitHash])
	}

	return result, nil
}

func resolveTagCommit(repo *git.Repository, hash plumbing.Hash) (plumbing.Hash, bool) {
	tag, err := repo.TagObject(hash)
	if err == nil {
		return resolveTagTarget(repo, tag.Target, 0)
	}

	if _, err := repo.CommitObject(hash); err == nil {
		return hash, true
	}

	return plumbing.Hash{}, false
}

func resolveTagTarget(repo *git.Repository, hash plumbing.Hash, depth int) (plumbing.Hash, bool) {
	if depth > 8 {
		return plumbing.Hash{}, false
	}

	if _, err := repo.CommitObject(hash); err == nil {
		return hash, true
	}

	tag, err := repo.TagObject(hash)
	if err != nil {
		return plumbing.Hash{}, false
	}

	return resolveTagTarget(repo, tag.Target, depth+1)
}

func isDirty(repo *git.Repository) (bool, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return false, ee.Wrap(err, "cannot get git worktree")
	}

	status, err := worktree.Status()
	if err != nil {
		return false, ee.Wrap(err, "cannot get git status")
	}

	return !status.IsClean(), nil
}

func describeToVersion(describe string) string {
	if describe == "" {
		return "0.0.0-unknown"
	}

	candidate := strings.TrimPrefix(describe, "v")
	if _, err := semver.NewVersion(candidate); err == nil {
		return candidate
	}

	return "0.0.0-" + sanitizePrerelease(describe)
}

var invalidPrereleaseChar = regexp.MustCompile(`[^0-9A-Za-z-]+`)

func sanitizePrerelease(value string) string {
	value = strings.TrimPrefix(value, "v")
	value = invalidPrereleaseChar.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "unknown"
	}

	return value
}

func renderEnv(info Info) string {
	entries := []struct {
		key   string
		value string
	}{
		{"VERSION", info.Version},
		{"GITCOMMIT", info.GitCommit},
		{"GITCOMMIT_SHORT", info.GitCommitShort},
		{"GITTAG", info.GitTag},
		{"GITDESCRIBE", info.GitDescribe},
		{"GITDIRTY", fmt.Sprintf("%t", info.GitDirty)},
		{"BUILDTIME", info.BuildTime},
		{"BUILDID", info.BuildID},
		{"KITTY_VERSION", info.KittyVersion},
		{"KITTY_COMMIT", info.KittyCommit},
		{"KITTY_BUILD_AT", info.KittyBuildAt},
	}

	var b strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&b, "%s=%s\n", entry.key, quoteEnvValue(entry.value))
	}

	return b.String()
}

func quoteEnvValue(value string) string {
	if isSafeEnvValue(value) {
		return value
	}

	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func isSafeEnvValue(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			strings.ContainsRune("_./:@%+-", r) {
			continue
		}

		return false
	}

	return true
}

func shortHash(hash string) string {
	if len(hash) <= 8 {
		return hash
	}

	return hash[:8]
}
