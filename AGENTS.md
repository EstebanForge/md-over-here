# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go CLI application (`md-over-here`).

- `cmd/md-over-here/`: CLI entrypoint, Cobra commands, and command-level tests.
- `internal/cache`, `internal/fetcher`, `internal/extractor`, `internal/converter`, `internal/processor`: core application packages.
- `scripts/checks.sh`: wrapper for the full local/CI validation flow.
- `.github/workflows/`: CI (`build.yml`) and release automation (`release.yml`).
- `bin/` and `dist/`: local build and release artifacts (generated).

Keep new runtime code in `internal/` unless it is part of the CLI surface.

## Build, Test, and Development Commands
Use the `Makefile` as the canonical interface:

- `make build`: build `bin/md-over-here`.
- `make test`: download/verify modules, run tests (with `-race` when available), and perform a build check.
- `make lint`: run `golangci-lint`.
- `make fmt`: run `gofmt` on all packages; applies `goimports` when installed.
- `make check` or `./scripts/checks.sh`: full gate (`fmt`, `vet`, `lint`, `test`, `build`).
- `make ci`: alias of `make check`.
- `make run ARGS='https://example.com'`: local CLI run.

## Coding Style & Naming Conventions
- Follow standard Go formatting; do not hand-format (`make fmt`).
- Use idiomatic Go naming: exported identifiers in `CamelCase`, internal helpers in `camelCase`.
- Keep package names short, lowercase, and singular (`cache`, `fetcher`, `processor`).
- Prefer small, focused files in the relevant `internal/<module>/` package.

## Testing Guidelines
- Use the Go `testing` package with table-driven tests where useful.
- Place tests next to code in `*_test.go` files (for example, `internal/processor/processor_test.go`).
- Name tests `TestXxx` and ensure clear failure messages.
- Run `make test` before opening a PR; run `make test-coverage` when validating broader changes.

## Commit & Pull Request Guidelines
- Follow Conventional Commit style used in history: `feat:`, `fix:`, `chore:`, `docs:`.
- Keep commit subjects concise and imperative (for example, `fix: handle nil cache in processor`).
- Include a short summary of behavior changes in every PR.
- Link the related issue when one exists.
- Include test evidence (commands run, such as `make check`).
- Update `CHANGELOG.md` for user-visible changes.
