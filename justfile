# Claude Code Open - Justfile
# Modern command runner with better ergonomics than make
# Install: https://github.com/casey/just

# Variables
binary_name := "cco"
version := "0.7.0"
build_dir := "build"
main_package := "."

# Default recipe (runs when you just type 'just')
default: fmt test build

# List all available recipes
@list:
    just --list

# Build the binary
build:
    go build -ldflags="-s -w -X 'github.com/Davincible/claude-code-open/cmd.Version={{version}}'" -o {{binary_name}} {{main_package}}

# Build binaries for all platforms
build-all: clean
    #!/usr/bin/env bash
    mkdir -p {{build_dir}}

    platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

    for platform in "${platforms[@]}"; do
        IFS='/' read -r goos goarch <<< "$platform"
        output="{{build_dir}}/{{binary_name}}-${goos}-${goarch}"

        if [ "$goos" = "windows" ]; then
            output="${output}.exe"
        fi

        echo "Building for $goos/$goarch..."
        GOOS="$goos" GOARCH="$goarch" go build \
            -ldflags="-s -w -X 'github.com/Davincible/claude-code-open/cmd.Version={{version}}'" \
            -o "$output" \
            {{main_package}}
    done

    echo "✓ Built binaries for all platforms"

# Run all tests
test:
    go test -v ./...

# Run tests with race detection
test-race:
    go test -v -race ./...

# Run tests with coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "✓ Coverage report generated: coverage.html"
    @go tool cover -func=coverage.out | tail -1

# Show coverage in browser
coverage-view: coverage
    @echo "Opening coverage report in browser..."
    @which xdg-open > /dev/null && xdg-open coverage.html || open coverage.html

# Format code
fmt:
    gofmt -s -w .
    @echo "✓ Code formatted"

# Run linter (requires golangci-lint)
lint:
    @which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
    golangci-lint run
    @echo "✓ Lint checks passed"

# Run linter and auto-fix issues
lint-fix:
    @which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
    golangci-lint run --fix
    @echo "✓ Lint issues fixed"

# Clean build artifacts
clean:
    go clean
    rm -f {{binary_name}}
    rm -rf {{build_dir}}
    rm -f coverage.out coverage.html
    @echo "✓ Build artifacts cleaned"

# Download and tidy dependencies
deps:
    go mod download
    go mod tidy
    @echo "✓ Dependencies updated"

# Verify dependencies
verify:
    go mod verify
    @echo "✓ Dependencies verified"

# Install binary to system
install: build
    sudo cp {{binary_name}} /usr/local/bin/{{binary_name}}
    @echo "✓ {{binary_name}} installed to /usr/local/bin"

# Remove binary from system
uninstall:
    sudo rm -f /usr/local/bin/{{binary_name}}
    @echo "✓ {{binary_name}} removed from /usr/local/bin"

# Run in development mode with auto-reload
dev:
    @which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
    @echo "🔥 Starting development server with hot reload..."
    air

# Build Docker image
docker-build:
    docker build -t claude-code-open:{{version}} .
    docker tag claude-code-open:{{version}} claude-code-open:latest
    @echo "✓ Docker image built: claude-code-open:{{version}}"

# Run Docker container
docker-run:
    docker run --rm -p 6970:6970 -v ~/.claude-code-open:/root/.claude-code-open claude-code-open:latest

# Create release build with checksums
release: clean fmt test build-all
    #!/usr/bin/env bash
    cd {{build_dir}}
    sha256sum * > checksums.txt
    echo "✓ Release {{version}} built successfully"
    echo "📦 Binaries available in {{build_dir}}/"
    echo ""
    echo "Checksums:"
    cat checksums.txt

# Run CI checks locally (same as GitHub Actions)
ci: fmt lint test-race
    @echo "✓ All CI checks passed"

# Update version in all files
update-version NEW_VERSION:
    #!/usr/bin/env bash
    echo "Updating version to {{NEW_VERSION}}..."
    sed -i.bak 's/version := ".*"/version := "{{NEW_VERSION}}"/' justfile
    sed -i.bak 's/Version    = ".*"/Version    = "{{NEW_VERSION}}"/' cmd/root.go
    rm -f justfile.bak cmd/root.go.bak
    echo "✓ Version updated to {{NEW_VERSION}}"
    echo "Don't forget to update README.md changelog!"

# Quick check before committing
pre-commit: fmt lint test
    @echo "✓ Ready to commit!"

# Start the CCO server
start:
    ./{{binary_name}} start

# Stop the CCO server
stop:
    ./{{binary_name}} stop

# Show CCO status
status:
    ./{{binary_name}} status

# Start the web UI
ui:
    ./{{binary_name}} ui

# List all models
models:
    ./{{binary_name}} models list

# List all plugins
plugins:
    ./{{binary_name}} plugins list

# Generate shell completion
completion SHELL:
    ./{{binary_name}} completion {{SHELL}}

# Show version info
version:
    @echo "CCO Version: {{version}}"
    @./{{binary_name}} --version 2>/dev/null || echo "(binary not built yet - run 'just build')"

# Initialize git hooks
git-hooks:
    #!/usr/bin/env bash
    echo "#!/bin/sh" > .git/hooks/pre-commit
    echo "just pre-commit" >> .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✓ Git hooks installed"
    echo "  Pre-commit hook will run: fmt, lint, test"

# Benchmark tests
bench:
    go test -bench=. -benchmem ./...

# Security scan with gosec
security:
    @which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
    gosec ./...
    @echo "✓ Security scan complete"

# Find TODO comments in code
todos:
    @echo "📋 TODO items in code:"
    @grep -r "TODO" --include="*.go" . || echo "No TODOs found!"

# Check for outdated dependencies
outdated:
    go list -u -m all

# Show project statistics
stats:
    @echo "📊 Project Statistics"
    @echo "===================="
    @echo "Go files: $(find . -name '*.go' -not -path './vendor/*' | wc -l)"
    @echo "Test files: $(find . -name '*_test.go' | wc -l)"
    @echo "Lines of code: $(find . -name '*.go' -not -path './vendor/*' -not -name '*_test.go' -exec cat {} \; | wc -l)"
    @echo "Test coverage: $(go test -cover ./... 2>/dev/null | grep 'coverage:' | tail -1 || echo 'Run just coverage')"

# Help - show all recipes with descriptions
help:
    @just --list
