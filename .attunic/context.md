# Project Context

Part of the project memory: human-readable memory of what this project is.

The reading and writing rules for this file and the rest of `.attunic/` are defined in [attunic-protocol.md](../.agents/attunic-protocol.md). The user fills this file in, by hand or by asking an agent to help; change it only on such a request and after the user explicitly confirms that change (see [attunic-protocol.md's Write Rules](../.agents/attunic-protocol.md#write-rules)).

## Abstract

A newcomer should be able to tell what the software does without knowing what it is built with. One paragraph that describes the domain, not the technology, in simple terms:
<!-- Project-specific / Abstract -->

TFunifiler is a cross-platform file-management command line tool. It helps users work with files in bulk, and manage files metadata using a local SQLite database.

## Technical Specifications

A brief overview of the ingredients developers of this project should know.

Write one bullet per ingredient, as `- <Ingredient>: <value>.`, such as primary language, frameworks, architecture style, build system, test runner, linter, databases, environment prerequisites.

Write `None` as the value of an ingredient the project lacks but an agent would look for, such as a build system, test runner, or linter.

<!-- Project-specific / Technical Specifications -->

- Primary language: Go.
- Build system: go.
- Test runner: go.
- Linter: gofmt.

## Layout

Layout of this project; [project-structure.md](../.agents/project-structure.md) holds the standard layouts of its languages and frameworks, kept for reference.

List of important folders, files and their roles in the project:

Write the list as a tree in a fenced block, one entry per line with its role as a `#` comment, mirroring the actual file system.

When the list is not empty and a top-level file/folder is missing from it, prompt the user to update it.

<!-- Project-specific / Layout -->

```
├── .agents/            # project AI agent instructions (attunic-protocol, coding-conventions, development-guideline, license-header, memory-template, project-structure, pull-request-guideline)
├── .attunic/           # project AI agent memory (context, memory, question)
├── .github/            # GitHub configuration as code
│   └── workflows/      # GitHub Actions release workflow
├── .vscode/            # Visual Studio Code configuration as code
├── app/                # wiring root of the program: cobra root, rpc server, rpc dispatch
├── config/             # global configuration from file, environment variables
├── core/               # project-level std lib
│   └── compression/    # file compression types and constants
├── crypto/             # crypto primitives
│   └── hasher/         # file/data hashers (crc32, md5, sha1, sha256,...)
├── db/                 # SQLite support via Gorm
├── diag/               # logging and flow/progress trackers
├── encoding/           # text codec helpers
│   └── checksum/       # checksum file readers/writers
├── engine/             # core of the program: cobra wiring, CLI modules and shared helpers
├── filesys/            # cross-platform filesystem ops
│   └── exec/           # wrappers for external tools
├── internal/           # non-importable internals
│   └── nullable/       # forked nullable types (bool, int), from github.com/tee8z/nullable
├── server/             # lightweight HTTP server for serving RPC
├── task/               # in-memory task queue for the RPC
├── tui/                # terminal UI components based on Bubbletea v2
├── xlib/               # small shared utility helpers
├── .copier-answers.yml # project Copier answers
├── .editorconfig       # EditorConfig configuration
├── .gitignore          # git ignore patterns
├── AGENTS.md           # project AI agent entry point
├── CLAUDE.md           # project Claude Code entry point
├── COPYING             # GPL-3.0 license text
├── README.md           # project overview
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── main.go             # program entry point
```

## Layer Dependencies

List of architecture layers and their relationships:

- Declare the innermost architecture layer first.
- An architecture layer may depend only on itself and on the architecture layers in its `Depends on` column, all declared before it.
- If an architecture layer appears in the table below, the dependency graph must be enforced.

| Architecture layer | Path | Depends on | Must never contain |
|--------------------|------|------------|--------------------|
| <!-- Project-specific / Layer Dependencies --> | | | |

### Deliberate Exceptions

List of documented exceptions that cannot follow the architecture layers above:

- Any place the rules above are knowingly broken, with the reason and the condition for fixing it.
- Only exceptions to the architecture layers above belong here.
- An undocumented exception is a defect; a documented one is allowed, and revisited once its `Revisit-if` condition holds.

| What | Why | Revisit-if |
|------|-----|------------|
| <!-- Project-specific / Deliberate Exceptions --> | | |

## Supporting Tools

All the supporting tools declared here are opinionated for the sake of consistent results across development environments.

List each required tool/library under its own heading, followed by the commands for each use case.

Common use cases include: Build, Test, Lint, Format, Install.

<!-- Project-specific / Supporting Tools -->

### Build

- `go build -o ./build/unifiler .` - build the program from the root package, output in `./build` (append `.exe` on Windows).

### Test

- `go test ./...` - run the whole test suite with the go test runner.
- `go test -race ./...` - run the suite with the race detector; required for changes that touch concurrency, per [development-guideline.md](../.agents/development-guideline.md).
- Narrower: `go test ./engine/...`, `go test ./db/...`, etc.

### Lint

- `gofmt -l .` - check formatting; must report no files.
- `gofmt -w .` - apply canonical formatting.
- `go vet ./...` - go static analysis; run on every change that touches concurrency.
- `goimports -l .` / `goimports -w .` - check/fix import grouping (stdlib, then third-party, then local).
- `go mod tidy -diff` / `go mod tidy` - check/fix that `go.mod` and `go.sum` match the imports; the check must report no diff.

## Project-specific

<!-- Project-specific / Project-specific -->
