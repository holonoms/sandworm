# Sandworm 🪱

A simple tool that concatenates a bunch of files (think a project folder)
into a single document, perfect for feeding to large language models and
"talk-to-your-codebase" workflows.

It integrates with Claude Projects and automates the filesync process, which
allows you to quickly iterate on your code with the help of Claude.

## Quick start

```bash
brew tap umwelt-studio/tap
brew install sandworm
brew sandworm --help
```

## What does it do?

Sandworm reads your project directory, combines all text files into a
single text file, and uploads this file to a configured Claude project. This
Allows you to chat with Claude having the most recent version of your entire
project context.

## What is good for?

The typical workflow is:

1. Make changes to your project
2. Run sandworm to update your Claude project
3. Ask Claude to iterate on the new code (e.g. build new feature, refactor code)
4. Apply Claude's suggested changes & run sandworm again

Rinse and repeat. Used in this fashion, sandworm allows you to very quickly add
new features to your software projects by leveraging Claude (or a similar LLM).

## How does it work?

Sandworm will initially look for a `.sandwormignore` file, falling back to a
`.gitignore` file if not found. This ignore file (which follows `.gitignore`
[inclusion/exclusion patterns](https://git-scm.com/docs/gitignore#_pattern_format))
is how sandworm determines which files to combine into the output file.

It then creates a project file that consists of a file & folder tree at the top,
followed by individual file's contents.

The first time you run `sandworm push` (or just `sandworm`), you'll be prompted
to configure the Claude Project you want to sync with. Once that's done,
subsequent runs of sandworm will take care of syncing the newly generated file
with the Project's file. Every new Project Chat with Claude will include your
whole project as context.

## Advanced usage

```
Sandworm v0.1.1 - Project file concatenator

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

Keep the uploaded file for inspection:

```bash
sandworm push -k
```

Generate only, don't push to Claude Project:

```bash
sandworm generate
```

### Output Format

The generated file will have the structure:

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
