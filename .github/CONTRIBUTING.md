# Contributing to Claude Code Open

Thank you for your interest in contributing to CCO! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- **Go 1.22+** - [Install Go](https://golang.org/doc/install)
- **just** - Modern command runner ([Install just](https://github.com/casey/just#installation))
- **golangci-lint** - Linter ([Install golangci-lint](https://golangci-lint.run/usage/install/))

### Quick Start

```bash
# Clone the repository
git clone https://github.com/Davincible/claude-code-open
cd claude-code-open

# Install dependencies
just deps

# Run tests
just test

# Build the binary
just build

# See all available commands
just list
```

## Development Workflow

### Using Just

We use `just` instead of `make` for a better developer experience. Here are the most common commands:

```bash
just build          # Build the binary
just test           # Run tests
just test-race      # Run tests with race detection
just coverage       # Generate coverage report
just lint           # Run linter
just fmt            # Format code
just pre-commit     # Run all checks before committing
just ci             # Run all CI checks locally
```

### Code Quality

Before submitting a PR, ensure:

1. **All tests pass**: `just test`
2. **Code is formatted**: `just fmt`
3. **Linter passes**: `just lint`
4. **Coverage is maintained**: `just coverage`

Or run everything at once:

```bash
just pre-commit
```

### Setting Up Git Hooks

Install pre-commit hooks to automatically check code before committing:

```bash
just git-hooks
```

This will run `fmt`, `lint`, and `test` automatically before each commit.

## Testing

### Writing Tests

- Place tests in `*_test.go` files next to the code being tested
- Use table-driven tests where appropriate
- Aim for high coverage (>80%)
- Test both success and error cases

### Running Tests

```bash
# Run all tests
just test

# Run with race detection
just test-race

# Generate coverage report
just coverage

# View coverage in browser
just coverage-view

# Run benchmarks
just bench
```

### Current Coverage

- **Config**: 71.7%
- **Providers**: 75.4%
- **Handlers**: 36.4%
- **Plugins**: 0% (new code, needs tests!)
- **Web UI**: 0% (new code, needs tests!)

**Help us improve coverage!** New features should include comprehensive tests.

## Code Style

### Go Conventions

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (automatic with `just fmt`)
- Keep functions small and focused
- Use meaningful variable names
- Add comments for exported functions and types

### Commit Messages

Use conventional commit format:

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat(providers): add Ollama provider for local models

Add support for running local models via Ollama, including:
- Request/response transformation
- Streaming support
- Comprehensive tests

Closes #123
```

```
fix(webui): correct token count display

The token counter was showing cached tokens incorrectly.
Now properly displays input + output tokens.
```

## Adding New Features

### Adding a New Provider

1. Create provider implementation in `internal/providers/`
2. Implement the `Provider` interface
3. Add comprehensive tests in `*_test.go`
4. Register provider in `internal/providers/registry.go`
5. Add default configuration in `internal/config/config.go`
6. Update README.md with provider documentation

Example structure:
```go
// internal/providers/newprovider.go
type NewProvider struct {
    Provider *config.Provider
}

func NewNewProvider(cfg *config.Provider) *NewProvider {
    return &NewProvider{Provider: cfg}
}

func (p *NewProvider) Name() string {
    return "newprovider"
}

// Implement remaining interface methods...
```

### Adding a New Plugin

1. Create plugin in `internal/plugins/builtin/`
2. Implement appropriate plugin interface(s)
3. Add configuration to `internal/config/config.go`
4. Add plugin loading in `cmd/start.go` and `cmd/ui.go`
5. Update `cmd/plugins.go` with new plugin info
6. Add tests

### Adding CLI Commands

1. Create command file in `cmd/`
2. Register command in `cmd/root.go`
3. Follow cobra command conventions
4. Add help text and examples
5. Update README.md documentation

## Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** with descriptive commits
3. **Add tests** for new functionality
4. **Run all checks**: `just pre-commit`
5. **Update documentation** (README, CHANGELOG, etc.)
6. **Push your branch** and create a PR
7. **Respond to review feedback** promptly

### PR Title Format

Use the same format as commit messages:
```
feat(scope): add feature description
fix(scope): fix bug description
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Motivation
Why is this change needed?

## Changes
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Added tests
- [ ] All tests pass
- [ ] Linter passes
- [ ] Coverage maintained/improved

## Documentation
- [ ] README updated
- [ ] Code comments added
- [ ] CHANGELOG updated (for releases)

## Screenshots (if applicable)
Add screenshots for UI changes
```

## Release Process

Releases are handled by maintainers:

1. Update version using `just update-version X.Y.Z`
2. Update CHANGELOG.md
3. Create and push version tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
4. GitHub Actions automatically builds and publishes release

## Project Structure

```
claude-code-open/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── start.go           # Start server
│   ├── ui.go              # Web UI
│   └── ...
├── internal/
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── plugins/           # Plugin system
│   │   └── builtin/       # Built-in plugins
│   ├── providers/         # LLM provider implementations
│   ├── server/            # HTTP server
│   └── webui/             # Web UI server
│       └── static/        # Web UI assets
├── tests/                 # Integration tests
├── .github/
│   └── workflows/         # CI/CD pipelines
├── justfile              # Command runner recipes
└── README.md             # Main documentation
```

## Getting Help

- **Documentation**: See [README.md](../README.md)
- **Issues**: [GitHub Issues](https://github.com/Davincible/claude-code-open/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Davincible/claude-code-open/discussions)

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
