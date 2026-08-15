# Contributing to Go-Charts

## Prerequisites

- Go 1.26+
- [golangci-lint](https://golangci-lint.run/)
- go-gui (sibling directory at `../go-gui`)

## Build and Test

Run the full local validation gate before pushing a branch:

```bash
make prepush
```

`make prepush` approximates the CI matrix from one host: race-enabled tests,
`go vet`, lint, and a build. It aborts on the first failing target. Individual
targets are available for a tighter loop while iterating:

```bash
make build       # build all packages
make test        # run all tests
make test-race   # tests with the race detector
make vet         # static analysis
make lint        # full lint, at the version CI pins
```

`make lint` first checks that golangci-lint is installed at the version CI pins
(`LINT_VERSION` in the Makefile, kept equal to the `version:` in
`.github/workflows/ci.yml`), so a local pass and a CI pass mean the same thing.

Gate targets run with `GOWORK=off`. A local `go.work` here points at
`../go-gui`, which CI never sees: CI checks out go-gui and go-glyph at the
`ref:` pinned in `ci.yml` and rewrites the replace directives to match. Those
refs are kept equal to the `require` versions in `go.mod`, so resolving `go.mod`
is exactly what CI validates.

### CI-only validation

- The OS matrix itself — CI runs on both `ubuntu-latest` and `windows-latest`,
  so a platform-specific failure on the OS you are not using can only be caught
  there.
- The gallery workflow, which renders the showcase under `xvfb` and needs a
  virtual X display.

## Coding Conventions

- **No variable shadowing.** Use `=` to reassign existing variables, not `:=`.
- **Clean lint and format.** All code must pass `golangci-lint run ./...` and
  `gofmt` with zero issues before committing.
- All chart types follow the `*Cfg` struct pattern from go-gui.
- Charts implement `gui.View` interface.

## Submitting Changes

1. Fork the repository and create a feature branch.
2. Make focused, single-purpose commits.
3. Add or update tests for any changed behavior.
4. Run the full check suite before pushing.
5. Open a pull request against `main`.

## License

Contributions are accepted under the
[PolyForm Noncommercial License 1.0.0](LICENSE).
