.PHONY: test test-race lint vuln tidy bench all

# Run the test suite.
test:
	go test ./...

# Run tests with the race detector and coverage, matching CI.
test-race:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run the linter.
lint:
	golangci-lint run ./...

# Check for known vulnerabilities, matching the security workflow.
vuln:
	govulncheck ./...

# Tidy module dependencies.
tidy:
	go mod tidy

# Run benchmarks.
bench:
	go test -bench=. -benchmem ./...

# Run the checks CI runs.
all: tidy lint test-race vuln
