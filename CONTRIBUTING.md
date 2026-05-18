# Contributing to apadana

Thank you for considering contributing! Here's how to get started.

## Development Setup

1. Clone the repo: `git clone https://github.com/PapaDanielVi/apadana.git`
2. Install Go 1.26+ (see [go.mod](go.mod) for minimum version)
3. Run tests: `go test ./...`
4. Run linter: `golangci-lint run --timeout=5m`

## Pull Request Process

1. Fork the repo and create a feature branch
2. Make your changes, add tests
3. Ensure `go test ./...` and `golangci-lint run` pass
4. Submit a PR with a clear description

## Code Style

- Follow standard Go conventions (run `goimports` or `gofmt`)
- Add tests for new functionality
- Update documentation/README if needed
- Use `t.Parallel()` in tests
- Use `package foo_test` for test packages (only use exported symbols)

## Reporting Issues

Use GitHub Issues to report bugs or request features. Include:
- Go version
- Expected vs actual behavior
- Minimal reproduction steps