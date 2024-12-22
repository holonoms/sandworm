# Sandworm 🪱

A simple tool that concatenates a bunch of files (think a project folder)
into a single document, perfect for feeding to large language models.
Sandworm preserves file structure context by:

1. Including a full directory tree at the start of the document
2. Clearly separating each file with headers containing their full paths
3. Respecting `.gitignore` rules to exclude unnecessary files

Perfect for when you need to:

- Upload your project context to ChatGPT, Claude, or other LLMs
- Create a single-file snapshot of your project
- Generate documentation that requires full codebase context

Sandworm was developed with Sandworm 🙃

## Installation

```bash
go get sandworm
```

## Basic usage

```bash
sandworm [directory]
```

This will:

1. Generate a directory tree of your project
2. Concatenate all files (respecting `.gitignore` or custom ignore file such as `.sandwormignore`)
3. Create a temporary `sandworm-<timestamp>.txt` in the current directory
4. Upload that file to the configured Claude project
5. Delete the temporary file

> [!NOTE]
> First run will prompt you to setup the tool for a specific Claude project

## Advanced usage

```
Sandworm vX.Y.Z - Project file concatenator

Usage: sandworm [command] [options] [directory]

Commands:
    generate    Generate concatenated file only
    push        Generate and push to Claude (default)
    purge       Remove all files from Claude project
    setup       Configure Claude project

Options:
  -ignore string
        Ignore file (default: .gitignore)
  -k    Keep the generated file (short flag)
  -keep
        Keep the generated file after pushing (only affects push)
  -o string
        Output file (short flag)
  -output string
        Output file (defaults to temp file on push and sandworm.txt for generate)
  -v    Show version (short flag)
  -version
        Show version
```

### Examples

Basic usage with default options:

```bash
# Use current directory as root folder
sandworm
```

Specify custom output file:

```bash
sandworm src/ -o context.txt
```

Use custom ignore file:

```bash
sandworm src/ --ignore custom-ignore.txt
```

### Output Format

The generated file will have this structure:

```
PROJECT STRUCTURE:
================

/
├── components
│   ├── Button.tsx
│   └── Card.tsx
└── pages
    └── index.tsx

FILE CONTENTS:
=============

================================================================================
FILE: components/Button.tsx
================================================================================
[file contents here]

================================================================================
FILE: components/Card.tsx
================================================================================
[file contents here]

...
```

### File Filtering

If a custom ignore file is provided, Sandworm will use only rules in that file.
Otherwise, it'll follow git ignore rules (`git ls-files`) and add a few extra
ignore rules to exclude binary files and other files that are typically checked
in but irrelevant in the context of LLM assistance.

## Development

```bash
# Setup tooling
asdf install (or something else that supports .tool-versions)

# Run project from sources
just run --help

# Build binary & run it
just build && bin/sandworm

# Run all checks
just lint

# Run tests
just test

# Other tasks
just --list
```

## Cutting a new release

New releases are produced with [goreleaser](.goreleaser.yml).
To create a new release, simply push a tag in the format `vX.Y.Z`.
A new release will be automatically
[created and uploaded to GitHub](./.github/workflows/release.yml).

```bash
git tag -a v0.42.0 -m "Release v0.42.0"
git push origin v0.42.0
```
