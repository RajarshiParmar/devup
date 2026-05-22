# devup

Spin up your local development environment in a single command.

`devup` automates the tedious startup sequence for macOS developers using OrbStack as their Docker runtime:

1. (Optionally) runs `brew update && brew upgrade`
2. Switches the Docker context to OrbStack
3. Starts OrbStack if Docker isn't running (with polling/timeout)
4. Runs `make start` in your project directory

## Installation

### From source

```sh
git clone https://github.com/RajarshiParmar/devup.git
cd devup
make install
```

This installs the binary to `~/bin/devup`. Make sure `~/bin` is in your `PATH`.

### With Go

```sh
go install github.com/RajarshiParmar/devup/cmd/devup@latest
```

## Prerequisites

- macOS
- [OrbStack](https://orbstack.dev/) installed
- `docker`, `make`, and `brew` available on `PATH`

Run `devup doctor` to verify your environment is ready.

## Usage

```sh
# Start everything with defaults
devup

# Specify a project directory
devup --dir ~/projects/myapp

# Include brew upgrade in the startup sequence
devup --upgrade

# Preview what would run without executing
devup --dry-run

# Check prerequisites
devup doctor

# Print version info
devup version
```

## Flags

| Flag | Default | Env Var | Description |
|------|---------|---------|-------------|
| `--dir` | `.` | `DEVUP_DIR` | Project directory for `make start` |
| `--timeout` | `60` | `DEVUP_TIMEOUT` | Max seconds to wait for Docker engine |
| `--upgrade` | `false` | `DEVUP_UPGRADE` | Run `brew update && brew upgrade` first |
| `--dry-run` | `false` | | Print actions without executing |
| `--no-color` | `false` | `NO_COLOR` | Disable colored output |
| `--verbose` | `false` | `DEVUP_VERBOSE` | Enable debug logging |

## Configuration

`devup` reads configuration from `$HOME/.config/devup/config.yaml`. Environment variables with the `DEVUP_` prefix override both the config file and default flag values.

## Development

```sh
make build            # Compile to ./bin/devup
make test             # Run tests with race detector + coverage
make lint             # Run golangci-lint
make doctor           # Build and run devup doctor
make tidy             # Tidy and verify go modules
make clean            # Remove build artifacts
make release-snapshot # Dry-run goreleaser without publishing
```

## License

See [LICENSE](LICENSE) for details.
