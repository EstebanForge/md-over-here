# md-over-here

> *"Get that content over here... as markdown!"* 🦂

A Go CLI tool that fetches web pages and converts them to clean markdown or structured data (TOON/JSON), optimized for feeding into Agents/LLMs with AXI (Agent eXperience Improvement) features.

## Features

- **Clean Content Extraction**: Uses Mozilla's Readability algorithm to extract main article content
- **Metadata Extraction**: Captures title, author, publish date, and description
- **Smart Caching**: Local file-based caching with 24-hour TTL at `~/.config/md-over-here/cache`
- **Multiple URLs**: Process multiple URLs in a single command
- **Flexible Output**: Write to stdout or file
- **Robust Error Handling**: Partial success - continues processing even if some URLs fail
- **AXI (Agent eXperience Improvement)**: Optimized for AI agents and LLMs
  - **TOON Format**: Token-efficient, structured text output (default format)
  - **Multiple Output Formats**: `toon` (default), `markdown`, `json`
  - **Field Selection**: Choose specific fields with `--fields` flag
  - **Content Truncation**: Control output length with `--truncate` and `--full`
  - **Batch Aggregates**: Show operation statistics with `--aggregates`
  - **Dashboard**: Run without arguments for an interactive dashboard
  - **Shell Hooks**: Integration via `hook init/status/uninstall` subcommands
- **Interface-Based Architecture**: Designed for future extensibility (headless Chrome support planned)

## Installation

### Homebrew (Linux, macOS and WSL)

```bash
brew install EstebanForge/tap/md-over-here
```

### Build from Source

```bash
git clone https://github.com/EstebanForge/md-over-here
cd md-over-here
make build
# outputs: bin/md-over-here
```

Or using Go directly:

```bash
go build -o bin/md-over-here ./cmd/md-over-here
```

### Install from Source to PATH

```bash
make install
```

Or using Go directly:

```bash
go install ./cmd/md-over-here
```

### Shorthand

On first run, the tool automatically creates a symlink at `~/.local/bin/mdoh`, allowing you to use the shorter `mdoh` command instead of `md-over-here`. This works for both Homebrew and manual installations.

Make sure `~/.local/bin` is in your PATH. If it's not already, add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Development

```bash
# Run all development checks (format, vet, lint, test, build)
make check
./scripts/checks.sh # wrapper around make check

# CI alias (same flow)
make ci

# Legacy coverage-heavy CI flow
make ci-coverage

# Build variants
make build          # unsigned local binary
make build-signed   # optional macOS signing

# Existing dev target
make dev

# Run individual commands
make test           # Run tests
make lint           # Run linter
make fmt            # Format code
make vet            # Run go vet
make test-coverage  # Run tests with coverage

# See all available commands
make help
```

## Usage

### Basic Usage

> **Note:** After first run, you can use `mdoh` as a shorthand for `md-over-here`.

```bash
# Single URL (outputs TOON format by default - optimized for agents/LLMs)
md-over-here https://example.com/article

# Output in markdown format
md-over-here --format markdown https://example.com/article

# Output in JSON format
md-over-here --format json https://example.com/article

# Select specific fields (TOON format only)
md-over-here --fields url,title,content https://example.com/article

# Truncate content to specified bytes
md-over-here --truncate 1000 https://example.com/article

# Show full content (bypass truncation)
md-over-here --full https://example.com/article

# Save to file (single or multiple URLs combined)
md-over-here --save article.md https://example.com/article

# Multiple URLs to stdout with aggregate statistics
md-over-here --aggregates https://example.com/article-1 https://example.com/article-2

# Save multiple URLs to one file
md-over-here --save combined.md https://example.com/article-1 https://example.com/article-2

# Bypass cache
md-over-here --no-cache https://example.com/article

# Verbose mode
md-over-here -v https://example.com/article

# Dashboard (run without arguments)
md-over-here

# Cache management
md-over-here cache stats    # Show cache statistics
md-over-here cache clear    # Clear all cached content

# Shell hooks (for advanced integration)
md-over-here hook init      # Install shell hooks
md-over-here hook status    # Show hook status
md-over-here hook uninstall # Remove shell hooks
```

### Available Flags

| Flag | Description |
|------|-------------|
| `-s, --save <file>` | Save to file (combines multiple URLs with separators) |
| `--format <type>` | Output format: `toon` (default), `markdown`, `json` |
| `--fields <list>` | Comma-separated fields to output (TOON format only) |
| `--truncate <bytes>` | Maximum content length in bytes (0 = no limit) |
| `--full` | Bypass truncation and show full content |
| `--aggregates` | Show aggregate statistics for batch operations |
| `--no-help` | Suppress help suggestions |
| `--no-cache` | Disable caching for this request |
| `--cache-dir <path>` | Custom cache directory (default: `~/.config/md-over-here/cache`) |
| `-v, --verbose` | Show metadata and cache status |
| `--timeout <duration>` | HTTP timeout (default: 30s) |
| `--user-agent <string>` | Custom User-Agent header |
| `-h, --help` | Show help message |

### Cache Subcommands

| Command | Description |
|---------|-------------|
| `md-over-here cache stats` | Display cache statistics (entries, size, location) |
| `md-over-here cache clear` | Remove all cached content |

Both cache subcommands support the `--cache-dir` flag to specify a custom cache directory.

## Output Formats

### TOON (Default)

TOON (Text-Oriented Object Notation) is a token-efficient, structured text format optimized for AI agents and LLMs. It provides clean, readable output with minimal tokens.

```bash
# TOON output (default)
md-over-here https://example.com/article
```

Output example:
```
url: https://example.com/article
title: Article Title
author: John Doe
published: 2025-01-15
description: Article description here
content: |
  # Article Title

  [Clean article content in markdown...]
fetched: 2026-01-03T16:47:00Z
cached: false
```

### Markdown

Traditional markdown output with metadata headers.

```bash
md-over-here --format markdown https://example.com/article
```

### JSON

Structured JSON output for programmatic processing.

```bash
md-over-here --format json https://example.com/article
```

### Field Selection

With TOON format, you can select specific fields to output:

```bash
# Only URL and title
md-over-here --fields url,title https://example.com/article

# Multiple fields
md-over-here --fields url,title,author,content https://example.com/article
```

Available fields: `url`, `title`, `author`, `published`, `description`, `content`, `fetched`, `cached`, `error`

### Content Truncation

Control output length for token efficiency:

```bash
# Limit to 1000 bytes
md-over-here --truncate 1000 https://example.com/article

# Show full content (bypass limits)
md-over-here --full https://example.com/article
```

### Batch Operations

Process multiple URLs with aggregate statistics:

```bash
# Show statistics
md-over-here --aggregates https://example.com/article-1 https://example.com/article-2
```

Aggregate output:
```
total: 2
success: 2
failed: 0
cached: 0
errors: 0
---
[results...]
```

## Output Format

```markdown
# Article Title

**URL:** https://example.com/article
**Author:** John Doe
**Published:** 2025-01-15
**Description:** Article description here

---

[Clean article content in markdown...]

---
<!-- Fetched: 2026-01-03T16:47:00Z -->
```

When processing multiple URLs, articles are separated by:

```markdown
---
## Next Article
---
```

## Caching

### Cache Location

Cached content is stored at `~/.config/md-over-here/cache/` by default.

### Cache Structure

```
~/.config/md-over-here/
├── cache/
│   ├── <sha256-hash>.json
│   └── <sha256-hash>.json
└── config.toml (future: user preferences)
```

### Cache Format

Each cached entry is stored as a JSON file:

```json
{
  "url": "https://example.com/article",
  "fetchedAt": "2026-01-03T16:46:53Z",
  "markdown": "# Article Title\n\n...",
  "metadata": {
    "Title": "Article Title",
    "Author": "John Doe",
    "PublishDate": "2025-01-15",
    "Description": "Article description"
  }
}
```

### Cache Key

- URLs are normalized (lowercase scheme/host, sorted query params, no fragment)
- SHA256 hash of normalized URL is used as cache filename
- TTL: 24 hours based on file modification time

### Cache Management

```bash
# Show cache statistics
md-over-here cache stats

# Clear all cached content
md-over-here cache clear

# Use custom cache directory
md-over-here cache stats --cache-dir /custom/path
md-over-here cache clear --cache-dir /custom/path
```

## Examples

### Agent-Optimized Output (TOON)

```bash
# Default TOON format - perfect for LLMs
md-over-here https://example.com/article

# Select specific fields to minimize tokens
md-over-here --fields url,title,content https://example.com/article

# Truncate for controlled input size
md-over-here --truncate 2000 https://example.com/article

# Batch with statistics
md-over-here --aggregates https://example.com/article-1 https://example.com/article-2
```

### Traditional Markdown Output

```bash
# Single URL
md-over-here --format markdown https://example.com/article

# Multiple URLs
md-over-here --format markdown https://example.com/article-1 https://example.com/article-2
```

### Structured JSON Output

```bash
# For programmatic processing
md-over-here --format json https://example.com/article

# With field selection
md-over-here --format json --fields url,title,author https://example.com/article
```

### Save to file

```bash
# Single URL to file
md-over-here --save article.md https://example.com/article

# Multiple URLs combined to one file
md-over-here --save research.md \
  https://blog.example.com/post-1 \
  https://blog.example.com/post-2 \
  https://blog.example.com/post-3
```

### Organize into subdirectories

```bash
# Parent directories are created automatically
md-over-here --save articles/2025/article.md https://example.com/article

# Absolute paths work too
md-over-here --save /path/to/docs/article.md https://example.com/article
```

### Process with custom timeout

```bash
md-over-here --timeout 60s https://slow-site.com/article
```

### Bypass cache for fresh content

```bash
md-over-here --no-cache https://news.example.com/breaking-story
```

### Dashboard and hooks

```bash
# Interactive dashboard
md-over-here

# Shell integration
md-over-here hook init
```

## Advanced Features

### Dashboard

Run `md-over-here` without arguments to launch an interactive dashboard showing cache statistics, recent activity, and quick actions.

```bash
md-over-here
```

### Shell Hooks

Install shell hooks for enhanced integration:

```bash
# Install hooks
md-over-here hook init

# Check hook status
md-over-here hook status

# Remove hooks
md-over-here hook uninstall
```

Hooks provide ambient integration with your shell for streamlined workflows.

## Architecture

### Design Principles

- **Agent-First Design**: AXI features optimize output for AI agents and LLMs
- **Interface-based Fetcher**: Designed to support future Chrome/headless browser backend
- **Graceful Degradation**: Falls back to full HTML if content extraction fails
- **Partial Success**: Processes all URLs even if some fail
- **Simple Caching**: JSON files for debuggability and simplicity
- **TOON Format**: Token-efficient structured output for agent consumption

## Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion
- [codeberg.org/readeck/go-readability/v2](https://codeberg.org/readeck/go-readability) - Content extraction

## Future Enhancements

### Planned

- **Headless Chrome Support**: `--use-chrome` flag for JS-heavy sites (SPAs, lazy-loaded content)
- **Parallel Processing**: `--parallel` flag for faster batch operations
- **Enhanced Hooks**: More shell integrations and custom hooks

### Under Consideration

- Image downloading and embedding
- Rate limiting for politeness
- Custom TOON field templates

## Troubleshooting

### Network Errors

```
Error: dial tcp: lookup example.com: no such host
```

**Solution**: Check network connectivity and DNS resolution

### HTTP 4xx/5xx Errors

```
Error: HTTP 404: 404 Not Found
```

**Solution**: Verify URL is correct and accessible. Tool continues with other URLs in batch.

### Content Extraction Failures

If Readability extraction fails, the tool falls back to converting the full HTML page to markdown.

### Cache Permission Errors

If cache directory creation fails, the tool continues without caching and shows a warning in verbose mode.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes in each version.

## License

MIT License - See [LICENSE](LICENSE) file for details
