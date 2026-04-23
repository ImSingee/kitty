---
name: kitty
description: Use Kitty to install and manage native Git hooks in a repository, configure hook commands, use Kitty extensions such as @lint-staged, set up lint-staged rules for staged or selected files, install hook-local tools, or troubleshoot Kitty hook installation and execution issues.
---

# Kitty

## Overview

Use Kitty as a lightweight Git hooks manager. Prefer it when a repository should commit hook definitions under `.kitty/` instead of relying on Node-specific hook managers.

## Workflow

1. Confirm the user is working at the Git repository root, or run commands with `kitty --root <repo-root> ...`.
2. Install Kitty if the `kitty` binary is missing.
3. Run `kitty install` at the repository root.
4. Add or update hook files with `kitty add <hook-name> "<command>"`.
5. Stage generated hook files with `git add .kitty/<hook-name>` and any config files the hook depends on.
6. Verify by running the hook command directly, then test through the relevant Git operation when practical.

## Install Kitty

Use one of these installation methods:

```sh
brew tap ImSingee/kitty
brew install kitty
```

```sh
go install github.com/ImSingee/kitty@latest
```

Downloaded release archives contain a single executable that can be placed on `PATH`.

## Initialize Hooks

Run:

```sh
kitty install
```

`kitty install` creates `.kitty/_/kitty.sh`, writes `.kitty/.gitignore`, and sets Git `core.hooksPath` to `.kitty`.

Important constraints:

- Run `kitty install` from the Git repository root. If the current directory is elsewhere, either `cd` to the root or use `kitty --root <repo-root> install`.
- Set `KITTY=0` to intentionally skip installation in environments where hooks should not be installed.
- Use `kitty install --no-tools` when hook-local tools should not be installed.
- Use `kitty install --direnv` to generate `.envrc` that prepends `.kitty/.bin` to `PATH` and reruns `kitty install --from-direnv`.
- Run `kitty uninstall` to unset Git `core.hooksPath`; it does not delete committed `.kitty/` files.

## Add Hooks

Use:

```sh
kitty add pre-commit "go test"
git add .kitty/pre-commit
```

`kitty add` creates `.kitty/<hook-name>` if missing. If the hook file already exists, it appends the command.

For extension commands, use the `@extension` shorthand:

```sh
kitty add pre-commit "@lint-staged"
git add .kitty/pre-commit
```

Kitty rewrites commands beginning with `@` to `kitty @...` inside the hook file.

Use the hidden `kitty set <hook-name> "<command>"` command only when replacing the entire hook file is intended. Prefer `kitty add` for normal setup because it preserves existing hook content.

## Lint-Staged

Use Kitty's built-in `@lint-staged` extension to run commands on selected Git files. The common pre-commit setup is:

```sh
kitty add pre-commit "@lint-staged"
git add .kitty/pre-commit
```

Run it manually with:

```sh
kitty @lint-staged
```

Target other file selections with:

```sh
kitty @lint-staged --status unstaged
kitty @lint-staged --status tracked
kitty @lint-staged --status changed
kitty @lint-staged --status all
```

Supported options include `--config`, `--diff`, `--diff-filter`, `--status`, `--stash`, `--shell`, `--verbose`, and `--allow-empty`.

## Lint-Staged Config

Place config in one of:

- `lint-staged` object inside `.kittyrc`, `.kittyrc.json`, or `kitty.config.json`
- `.lintstagedrc`
- `.lintstagedrc.json`

Do not assume JavaScript config works; `lint-staged.config.js` and `.lintstagedrc.js` are noted as coming soon in this project.

Kitty discovers lint-staged config files from Git-known files. Stage or commit new config files before expecting automatic discovery:

```sh
git add .lintstagedrc.json
```

Config can be either direct rules:

```json
{
  "*": "your-cmd"
}
```

or a `files` object:

```json
{
  "files": {
    "*": "your-cmd"
  }
}
```

Commands receive paths relative to the directory containing the selected config file:

```json
{
  "*.go": "go test ./..."
}
```

Use command modifiers when needed:

- `[absolute] cmd` passes absolute file paths.
- `[dir] cmd` passes matching directories instead of files.
- `[noArgs] cmd` runs without file arguments.
- `[prepend value] cmd` prepends a value to each file argument.

Only one lint-staged config file may exist in the same directory. Multiple config files in different directories are allowed; Kitty uses the closest config for each file.

## Tool Installation

`kitty install` runs `kitty tools install` unless `--no-tools` is provided. Use `.kitty/.bin` for hook-local tool binaries and prefer committing tool intent through Kitty-managed project files rather than relying on a developer's global `PATH`.

If a hook cannot find a tool, check whether `.kitty/.bin` is on `PATH`. With direnv, `kitty install --direnv` creates an `.envrc` that exports:

```sh
export PATH="$(pwd)/.kitty/.bin:$PATH"
```

## Troubleshooting

If `kitty add` reports `.kitty` is missing, run `kitty install` first.

If `kitty install` says to go to the repository root, change to the Git top-level directory and rerun it.

If hooks do not run, inspect:

```sh
git config core.hooksPath
ls -la .kitty
```

The hooks path should be `.kitty`, and hook files should be executable.

If lint-staged finds no config, verify the config filename is supported and already tracked or staged by Git.

If a lint-staged task needs repository-root paths instead of config-directory-relative paths, use `[absolute]`.

## Verification

Before reporting completion:

- Run `kitty install` or the exact changed Kitty command in a temporary or target repository.
- Inspect generated `.kitty/<hook-name>` when adding hooks.
- Stage required hook and config files.
- Run the underlying command directly.
- Trigger the hook through Git when practical, for example with a test commit in a disposable repository.
