# Repository Guidelines

## Project Structure & Module Organization

`crv` is a Go CLI for browsing JaCoCo XML in the terminal.

- `cmd/crv/main.go`: CLI entrypoint.
- `internal/app`: application orchestration.
- `internal/cli`: option parsing and validation.
- `internal/jacoco`: JaCoCo XML model/parser and performance tests.
- `internal/reportpath`: report path auto-detection (including `pom.xml` parsing).
- `internal/tui`: Bubble Tea-based interactive UI.
- `docs/`: requirements, test cases, verification, release notes.
- `bin/crv`: local built artifact (do not rely on this in commits).

## Build, Test, and Development Commands

- `go run ./cmd/crv --help`: run the CLI locally.
- `go build ./cmd/crv`: compile the binary for current OS/arch.
- `go test ./...`: run all unit/integration tests.
- `go test ./internal/jacoco -run TestParsePerformance1000Classes -count=1 -v`:
  run the performance check used in verification docs.
- `GOOS=linux GOARCH=amd64 go build -o /tmp/crv-linux-amd64 ./cmd/crv`: cross-build example.
- `markdownlint-cli2 "**/*.md"`: lint Markdown docs.

## Coding Style & Naming Conventions

- Follow standard Go formatting: run `gofmt` (or editor auto-format) on changed Go files.
- Keep package names short, lowercase, and purpose-based (`jacoco`, `reportpath`, `tui`).
- Use table-driven tests where practical; name tests as `Test<Behavior>`.
- Prefer explicit, user-facing error messages in CLI/TUI paths.

## Testing Guidelines

- Place tests next to implementation files (`*_test.go`) under the same package path.
- Include coverage for parsing, path detection, CLI option handling, and navigation behavior.
- Update `docs/TESTCASES.md`, `docs/TRACEABILITY.md`, and `docs/VERIFICATION.md` when behavior or acceptance evidence changes.

## Commit & Pull Request Guidelines

- Current history is minimal (`Initial commit`); use concise, imperative commit subjects going forward.
- Recommended format: `type(scope): summary` (e.g., `feat(tui): add sort toggle hint`).
- PRs should include:
- purpose and scope.
- linked issue/task ID.
- test evidence (`go test ./...` output summary).
- screenshots or terminal captures for TUI-visible changes.
