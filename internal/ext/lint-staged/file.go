package lintstaged

import (
	"path/filepath"

	"github.com/ImSingee/go-ex/mr"
)

// File
//
// it's underlying data is the path relative to the git root
type File struct {
	gitRelativePath       string
	relativePathToGitRoot string
	absolutePath          string
}

func NewFile(ctx *State, gitRelativePath string) *File {
	relativePathToGitRoot := gitRelativePath // TODO convert from git path (unix style) to system path (os style)
	absolutePath := normalizePath(filepath.Join(ctx.gitRoot, relativePathToGitRoot))

	return &File{
		gitRelativePath:       gitRelativePath,
		relativePathToGitRoot: relativePathToGitRoot,
		absolutePath:          absolutePath,
	}
}

func (f File) GitRelativePath() string {
	return f.gitRelativePath
}

func (f File) RelativePathToGitRoot() string {
	return f.relativePathToGitRoot
}

func (f File) AbsolutePath() string {
	return f.absolutePath
}

// Files is a slice of File
type Files []*File

func (files Files) GitRelativePaths() []string {
	return mr.Map(files, func(in *File, _index int) string {
		return in.GitRelativePath()
	})
}

func (files Files) RelativePathsToGitRoot() []string {
	return files.GitRelativePaths()
}

func (files Files) AbsolutePaths() []string {
	return mr.Map(files, func(in *File, _index int) string {
		return in.AbsolutePath()
	})
}

func NewFiles(ctx *State, relativePathsToGitRoot []string) Files {
	return Files(mr.Map(relativePathsToGitRoot, func(in string, _index int) *File {
		return NewFile(ctx, in)
	}))
}
