# ── TMS Backend Makefile ──────────────────────────────────────────────────────
# All commands run from tms-backend/.

.PHONY: build vet test test-race test-integration swagger lint tidy clean

# Build the binary (verifies the whole module compiles).
build:
	go build ./...

# Static analysis — should be clean before every commit.
vet:
	go vet ./...

# Unit tests with race detector.
test:
	go test -race -count=1 ./...

# Same as test but more verbose — useful in CI logs.
test-race:
	go test -race -v -count=1 ./...

# Integration tests (require running Postgres/Redis/ES).
# Mark integration test files with: //go:build integration
test-integration:
	go test -race -tags=integration -count=1 ./...

# Regenerate Swagger docs from annotations.
swagger:
	swag init -g main.go --output docs/ --parseDependency --parseInternal

# Lint with golangci-lint (install separately: https://golangci-lint.run/usage/install/).
lint:
	golangci-lint run ./...

# Tidy module dependencies.
tidy:
	go mod tidy

# Remove the compiled binary.
clean:
	go clean
	rm -f tms-backend tms-backend.exe
