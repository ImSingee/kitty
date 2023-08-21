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
kitty add .kitty/pre-commit "go test"
git add .kitty/pre-commit
```

Make a commit:

```sh
git commit -m "Keep calm and commit"
# `go test` will run
```

## Credits

- [husky](https://github.com/typicode/husky/tree/main)
- [lint-staged](https://github.com/okonet/lint-staged)

## License

MIT

