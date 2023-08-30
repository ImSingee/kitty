# kitty

> Modern native Git hooks made easy

Kitty improves your commits!

## Install

> Currently, kitty is only available for manual installation.
 
```shell
go install github.com/ImSingee/kitty@latest
``` 

## Usage

After cloning a project:

```shell
kitty install
```

Add a hook:

```shell
kitty add pre-commit "go test"
git add .kitty/pre-commit
```

Make a commit:

```sh
git commit -m "Keep calm and commit"
# `go test` will run
```

## Config

## Extension: lint-staged

kitty ships extension `lint-staged` to allow you to run commands on staged files.

In most cases, you can use `lint-staged` in the hook `pre-commit`:

```shell
kitty add pre-commit '@lint-staged'
git add .kitty/pre-commit
```

Or you can manually run

```shell
kitty lint-staged
```

### Configuration

*lint-staged* can be configured in many ways:

- `lint-staged` object in your kitty config
- `.lintstagedrc` or `.lintstagedrc.json` file (in JSON format)
- `lint-staged.config.js` or `.lintstagedrc.js` file (Comping Soon)

You can also place multiple configuration files in different directories inside a project. For a given staged file, the closest configuration file will always be used. But you can't have multiple configuration files in the same directory.

Configuration can be an object in two formats:

```json
{
  "files": {
    "*": "your-cmd"
  }
}
```

or simply (if you do not want to specify any advanced options):

```json
{
  "*": "your-cmd"
}
```

Inside the `files` object in format 1 or for whole object of format 2, each value is a command to run and its key is a glob pattern to use for this command.


`.kittyrc.json` example:

```json
{
  "lint-staged": {
    "*": "your-cmd"
  }
}
```

`.lintstagedrc.json` example:

```json
{
  "*": "your-cmd"
}
```

This config will execute `your-cmd` with the list of currently staged files passed as arguments.

So, considering you did `git add file1.ext file2.ext`, `lint-staged` will run the following command:

```shell
your-cmd file1.ext file2.ext
```

> **Note**
> Apart from node.js `lint-staged`, we do not pass absolute paths to the commands. Instead, we pass the relative path to the working directory (where lint-staged config is placed) to the command.
> If you want we pass absolute paths to the commands, you can prepend `[absolute] ` (note: space is required) before the command.
> 
> Example:
> ```json
> {
>   "*": "[absolute] your-cmd"
> }
> ```
> 
> Then we will run `your-cmd /absolute/path/to/file1.ext /absolute/path/to/file2.ext`

### Concurrency

We do not run commands in parallels now, but we plan to support it (and make it default) in the future.

### Platform Support

The `lint-staged` extension is not supported on windows platform now, please use it on Linux or macOS.

We are working on it, but it's not an urgent task. If you really need it, PR is welcome.

## Credits

- [husky](https://github.com/typicode/husky/tree/main)
- [lint-staged](https://github.com/okonet/lint-staged)

## License

MIT

