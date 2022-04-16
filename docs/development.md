# Development

## Tooling

- [asdf](https://asdf-vm.com/) - Extendable version manager with support for Ruby, Node.js, Erlang & more
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) - Updates your Go import lines, adding missing ones and removing unreferenced ones. Also formats your code in the same style as `gofmt`.
- [gotestsum](https://github.com/gotestyourself/gotestsum) - 'go test' runner with output optimized for humans, JUnit XML for CI integration, and a summary of the test results.
- [revive](https://github.com/mgechev/revive) - Fast, configurable, extensible, flexible, and beautiful linter for Go.
- [pre-commit](https://pre-commit.com/) - Framework for managing multi-language pre-commit hooks

## Setup local environment

### OS X

```shell
make setup-osx-env
```

This uses [Homebrew](https://brew.sh/) to install the `revive`, `asdf`, and `pre-commit` tools. It then uses `asdf` to
install the version of Go specified in `.tool-versions`, downloads the `goimports` and `gotestsum` tools, and sets up
the pre-commit hooks configured in `.pre-commit-config.yml`.

## GitHub pre-commit hooks

This project uses [pre-commit](https://pre-commit.com/) to run lint, vet, and unit check tests before you commit any
code.


## Continuous Integration
This project uses [GitHub Actions](https://github.com/sprak3000/go-glitch/actions) for build and continuous integration.

## Code Quality

This project uses [CodeClimate](https://codeclimate.com/github/sprak3000/go-glitch) to track code quality metrics and
trends.
