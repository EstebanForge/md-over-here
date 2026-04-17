# Changelog

All notable changes to md-over-here will be documented in this file.

<!-- RELEASE:START 1.0.0 -->
## [1.0.0] - 2026-04-17

### Added
- AXI (Agent eXperience Improvement) features for LLM/agent optimization
- TOON output format - token-efficient structured text (default for stdout)
- Universal TOON output: All CLI outputs (errors, dashboard, hooks, cache) use TOON format for LLM parsing
- Smart output format defaults: TOON for stdout, markdown for file saves
- Field selection with `--fields` flag
- Content truncation with `--truncate` and `--full` flags
- Batch aggregate statistics with `--aggregates` flag
- Simplified dashboard showing core commands in TOON format (run without arguments)
- Shell hooks for integration (`hook init/status/uninstall`)
- JSON output format option
- `--no-help` flag to suppress help suggestions
- `--version` flag to show version in TOON format
- TOON-formatted help output (both `--help` flag and `help` command)
- `--human` flag for help to show human-readable format when needed
- Support for `map[string]string` in TOON encoder for nested structures
- Simplified dashboard to 4 core commands (removed cache/hook commands)

### Changed
- Removed dead code and runtime overhead
- Enhanced build system with lint pinning, build-release, and CI parity
- Standardized release workflow and changelog markers
- Removed auto-triggers from build workflow (manual only)
- Added repository contributor guidelines
- Updated README for standardized check and build flows
- Standardized CI target and kept coverage variant
- Aligned fmt target with optional goimports step
- Standardized make checks, bin output, and signing hooks

<!-- RELEASE:END 0.2.0 -->

<!-- RELEASE:START 0.1.1 -->
## [0.1.1] - 2026-01-24

### Fixed
- Fixed Homebrew formula installation by creating platform-specific tarballs with generic binary names

<!-- RELEASE:END 0.1.1 -->

<!-- RELEASE:START 0.1.0 -->
## [0.1.0] - 2026-01-03

### Added
- Initial release.
<!-- RELEASE:END 0.1.0 -->
