.PHONY: test test-race vet lint lint-pin build prepush

# golangci-lint version pinned by CI (.github/workflows/ci.yml).
LINT_VERSION := v2.12.2

# Gate recipes resolve modules from go.mod, not from a go.work workspace.
# A local go.work here points at ../go-gui, which CI never sees: CI checks
# out go-gui and go-glyph at the `ref:` pinned in ci.yml and rewrites the
# replace directives to match. Those refs are kept equal to the require
# versions in go.mod, so resolving go.mod is exactly what CI validates.
# A workspace build would answer a different question.
GO := GOWORK=off go

# Run the test suite. Mirrors CI's `go test ./...` step.
test:
	$(GO) test ./...

# Race-enabled tests. CI does not run -race; this is a deliberate strict
# superset, cheap enough to be worth catching locally.
test-race:
	$(GO) test -race -count=1 ./...

# Static analysis. Mirrors CI's `go vet ./...` step.
vet:
	$(GO) vet ./...

# Verify golangci-lint is installed at the version CI pins, so a local pass
# and a CI pass mean the same thing.
lint-pin:
	@golangci-lint --version | grep -q "$(LINT_VERSION:v%=%)" || \
	  { echo "::error::golangci-lint $(LINT_VERSION) required. Run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)"; exit 1; }

lint: lint-pin
	golangci-lint run ./...

build:
	$(GO) build ./...

# Recommended full local validation before pushing (issue go-gui#314).
# Approximates the CI matrix from one host: race tests, vet, lint, build.
# Aborts on the first failing target.
#
# Omissions vs CI, by design:
#   - the OS matrix itself (CI runs ubuntu-latest and windows-latest)
#   - the gallery workflow, which renders the showcase under xvfb and
#     needs a virtual X display
prepush: test-race vet lint build
