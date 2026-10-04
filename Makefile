.PHONY: test test-race vet lint lint-bin build prepush

# Repo-local bin for the pinned linter. The pinned VERSION itself lives in
# tools/lint/go.mod -- see the $(LINT_BIN) rule below. `make lint` and CI
# both build from that file, so a local pass and a CI pass run one version.
LINT_DIR = $(CURDIR)/.bin
LINT_BIN = $(LINT_DIR)/golangci-lint

# Gate recipes resolve modules from go.mod, not from a go.work workspace.
# A local go.work here points at ../go-gui, which CI never sees: CI checks
# out go-gui and go-glyph at the `ref:` pinned in ci.yml and rewrites the
# replace directives to match. Those refs are kept equal to the require
# versions in go.mod, so resolving go.mod is exactly what CI validates.
# A workspace build would answer a different question.
GO := GOWORK=off go

# golangci-lint is its own binary, so $(GO) does not cover it — but it
# honours go.work the same way the toolchain does. Without GOWORK=off it
# type-checks against the ../go-gui working copy and reports breakage that
# CI, which builds the pinned versions, will never see.
LINT := GOWORK=off $(LINT_BIN)

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

# Build the pinned golangci-lint into .bin/. It rebuilds only when
# tools/lint/go.mod or go.sum change. GOWORK=off keeps a local go.work out
# of the build. GOOS/GOARCH/CGO_ENABLED are cleared so a caller that sets
# them to pick a lint target does not cross-compile the linter itself into
# a binary this host cannot run.
$(LINT_BIN): tools/lint/go.mod tools/lint/go.sum
	GOWORK=off GOOS= GOARCH= CGO_ENABLED=0 GOFLAGS= GOBIN=$(LINT_DIR) \
	  go -C tools/lint install \
	  github.com/golangci/golangci-lint/v2/cmd/golangci-lint

lint-bin: $(LINT_BIN)

lint: $(LINT_BIN)
	$(LINT) run ./...

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
