package versionext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestCollectTaggedRepository(t *testing.T) {
	disableZeaburGitEnv(t)

	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	commit := commitFile(t, repo, dir, "README.md", "hello\n", "initial")
	if _, err := repo.CreateTag("v1.2.3", commit, nil); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 5, 20, 6, 0, 0, 0, time.UTC)
	info, err := Collect(dir, now)
	if err != nil {
		t.Fatal(err)
	}

	if info.Version != "1.2.3" {
		t.Fatalf("Version = %q, want 1.2.3", info.Version)
	}
	if info.GitCommit != commit.String() {
		t.Fatalf("GitCommit = %q, want %q", info.GitCommit, commit.String())
	}
	if info.GitCommitShort != commit.String()[:8] {
		t.Fatalf("GitCommitShort = %q, want %q", info.GitCommitShort, commit.String()[:8])
	}
	if info.GitTag != "v1.2.3" {
		t.Fatalf("GitTag = %q, want v1.2.3", info.GitTag)
	}
	if info.GitDescribe != "v1.2.3" {
		t.Fatalf("GitDescribe = %q, want v1.2.3", info.GitDescribe)
	}
	if info.GitDirty {
		t.Fatal("GitDirty = true, want false")
	}
	if info.BuildTime != "2026-05-20T06:00:00Z" {
		t.Fatalf("BuildTime = %q", info.BuildTime)
	}
	if info.BuildID != "20260520060000-"+commit.String()[:8] {
		t.Fatalf("BuildID = %q", info.BuildID)
	}
}

func TestCollectDirtyRepositoryAfterTag(t *testing.T) {
	disableZeaburGitEnv(t)

	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	first := commitFile(t, repo, dir, "README.md", "hello\n", "initial")
	if _, err := repo.CreateTag("v1.2.3", first, nil); err != nil {
		t.Fatal(err)
	}
	second := commitFile(t, repo, dir, "main.go", "package main\n", "second")
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Collect(dir, time.Date(2026, 5, 20, 6, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	wantDescribe := "v1.2.3-1-g" + second.String()[:8] + "-dirty"
	if info.GitDescribe != wantDescribe {
		t.Fatalf("GitDescribe = %q, want %q", info.GitDescribe, wantDescribe)
	}
	if !info.GitDirty {
		t.Fatal("GitDirty = false, want true")
	}
}

func TestCollectUsesZeaburCommitWithoutGitRepository(t *testing.T) {
	t.Setenv(zeaburGitCommitSHAEnv, "abcdef1234567890abcdef1234567890abcdef12")

	dir := t.TempDir()
	now := time.Date(2026, 5, 20, 6, 0, 0, 0, time.UTC)
	info, err := Collect(dir, now)
	if err != nil {
		t.Fatal(err)
	}

	assertZeaburInfo(t, info, "abcdef1234567890abcdef1234567890abcdef12", now)
}

func TestCollectPrefersZeaburCommitOverGitRepository(t *testing.T) {
	t.Setenv(zeaburGitCommitSHAEnv, " fedcba9876543210fedcba9876543210fedcba98 ")

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 5, 20, 6, 0, 0, 0, time.UTC)
	info, err := Collect(dir, now)
	if err != nil {
		t.Fatal(err)
	}

	assertZeaburInfo(t, info, "fedcba9876543210fedcba9876543210fedcba98", now)
}

func TestRenderEnvQuotesOnlyWhenNeeded(t *testing.T) {
	info := Info{
		Version:        "0.0.0-test",
		GitCommit:      "abc",
		GitCommitShort: "abc",
		GitTag:         "v0.0.0",
		GitDescribe:    "v0.0.0-dirty",
		GitDirty:       true,
		BuildTime:      "2026-05-20T06:00:00Z",
		BuildID:        "20260520060000-abc",
		KittyVersion:   "DEV",
		KittyCommit:    "abc def",
	}

	got := renderEnv(info)
	for _, want := range []string{
		"VERSION=0.0.0-test",
		"GITCOMMIT=abc",
		"GITDIRTY=true",
		"KITTY_COMMIT='abc def'",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("env output does not contain %q:\n%s", want, got)
		}
	}
}

func disableZeaburGitEnv(t *testing.T) {
	t.Helper()

	t.Setenv(zeaburGitCommitSHAEnv, "")
}

func assertZeaburInfo(t *testing.T, info Info, commit string, now time.Time) {
	t.Helper()

	commitShort := commit[:8]
	if info.Version != "0.0.0-"+commitShort {
		t.Fatalf("Version = %q, want %q", info.Version, "0.0.0-"+commitShort)
	}
	if info.GitCommit != commit {
		t.Fatalf("GitCommit = %q, want %q", info.GitCommit, commit)
	}
	if info.GitCommitShort != commitShort {
		t.Fatalf("GitCommitShort = %q, want %q", info.GitCommitShort, commitShort)
	}
	if info.GitTag != "" {
		t.Fatalf("GitTag = %q, want empty", info.GitTag)
	}
	if info.GitDescribe != commitShort {
		t.Fatalf("GitDescribe = %q, want %q", info.GitDescribe, commitShort)
	}
	if info.GitDirty {
		t.Fatal("GitDirty = true, want false")
	}
	if info.BuildTime != now.Format(time.RFC3339) {
		t.Fatalf("BuildTime = %q, want %q", info.BuildTime, now.Format(time.RFC3339))
	}
	if info.BuildID != now.Format("20060102150405")+"-"+commitShort {
		t.Fatalf("BuildID = %q, want %q", info.BuildID, now.Format("20060102150405")+"-"+commitShort)
	}
}

func commitFile(t *testing.T, repo *gogit.Repository, dir, name, content, message string) plumbing.Hash {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worktree.Add(name); err != nil {
		t.Fatal(err)
	}
	hash, err := worktree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "kitty",
			Email: "kitty@example.com",
			When:  time.Date(2026, 5, 20, 6, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return hash
}
