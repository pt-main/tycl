# Changelog

All notable changes to **TYCL** (a typed configuration language for Go).

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versioning follows [SemVer](https://semver.org/).

[Russian version](changelog-ru.md)

## [1.4.0] - 2026-09-29

A release about observability and a complete CLI. The main work is a flat
diagnostics layer with human readable output, a stable JSON envelope and exact
source positions, plus a move to current `lc` and `tap` releases.

### Breaking

- **`lc` v1 -> v2.** The module path now follows semantic import versioning and
  is `github.com/pt-main/lc/v2`; every import, including subpackages, must end
  in `/v2`. `github.com/pt-main/lc/engine/core` becomes
  `github.com/pt-main/lc/v2/engine/core`. Three `parser3` call sites were moved
  to the new API: the `ParseError.Code` and `GrammarError.Code` fields were
  renamed to `Phase`, and the string literal `"Expect"` in `parseHint` was
  replaced with the `parser3.PhaseExpect` constant. A reference migration
  script ships with Lc: `docs/to-2.0-migrate.sh`.
- **`tap` -> `tap/go`.** Imports move from `github.com/pt-main/tap/...` to
  `github.com/pt-main/tap/go/...`. The Go implementation of tap now lives in
  the nested `tap/go` module, while the root `tap` module has been rewritten in
  Rust since 1.5.0 and contains no Go packages. The import name stays `tap`, so
  aliases need no changes.
- **The `tycl file` command was removed.** Its subcommands became standalone
  commands with their own argument parsing: `get`, `set`, `remove`,
  `structure`. The reason is that `file` had a private `--path` flag and a
  private syntax, so scripts and editor completions worked with different
  spellings of the same command. Migration:
  `tycl file get --path=a.tycl int port` -> `tycl get a.tycl port`. The full
  table is in the [Migration](#migration) section.
- The root file `contract.go` was removed; contracts live in the `contract`
  package.
- `test/test.go` was removed; coverage moved to package local `_test.go` files
  next to the code they test.
- Binary files under `build/` were deleted from the repository. Binaries do not
  belong in history and are produced by `build.sh`.

### Added

- **The `diag` package** - a flat diagnostics layer. Any `lc` or `tycl` error is
  expanded into a list of structures instead of a chain: `diag.Diagnostic`
  carries a stable code, a severity, a message, a hint, a `*Span` with the file
  position, the config path and the array index.
- **Human readable rendering with a caret.** `diag.Render` prints the source
  line, its number and `^^^^` under the offending token. No separate code path
  is needed: colours are resolved once at the end, so `--no-color` and
  redirected output behave identically.
- **A stable JSON envelope.** The `--json` flag makes every command print a
  single `{ok, command, diagnostics, data}` envelope to stdout and leave stderr
  empty, so the result can be parsed without filtering.
- **Human readable hints.** `diag` translates parser codes into advice: a
  missing comma before `}`, a comment before an opening `{`, a missing colon
  before a type.
- **A machine readable syntax tree.** The command `tycl ast` returns every node
  with its type, raw text and span. This is the language contract for editors
  and third party tools.
- **A contract layer for validation without the CLI.** `diag.CheckContract` and
  `tycl.Validate` return a ready list of diagnostics, and `tycl.ProcessSource`
  labels them with a file name.
- **stdin/stdout support.** The path `-` reads from stdin and writes to stdout,
  so `cat app.tycl | tycl gen - - json | jq .port` works as an ordinary pipe.
- **Predictable exit codes.** 0 success, 1 bad data, 2 usage error, 3 I/O
  error. The codes are covered by `cli_test.go`.
- New commands `query` (read several paths or all of them at once), `ast`,
  `docs`, `types`, `merge` (later files win, nested objects are merged by key)
  and `version`.
- The `--verbose` and `--debug` flags enable command dispatch tracing from tap.
  `--no-color` is accepted in both spellings: with a hyphen (the documented
  one) and with an underscore (the one tap itself checks).
- Tests for path parsing, config merging, values and typing: `lang/path.go`,
  `lang/merge.go`, `lang/values.go`, `lang/typing.go`.
- 547 lines of CLI tests, 166 plus 56 for diagnostics, 245 plus 11 for the
  language, 82 for contracts and 64 for formatting.

### Changed

- **Dependencies moved to current releases.** `lc` was updated from 1.5.4 to
  `github.com/pt-main/lc/v2` v2.0.1 and `tap` from 1.4.8 to
  `github.com/pt-main/tap/go` v1.5.8. The indirect dependency on
  `github.com/pt-main/tap` v1.4.7, which was pulled in by `lc` 1.5.x, is no
  longer needed.
- `go.sum` was rebuilt: entries for dead `lc` versions (1.5.0-pre, 1.5.2-1.5.4)
  and `tap` (1.4.7, 1.4.8, 1.5.6) were removed.
- Error message formatting moved into a separate `format` package with
  separators that point at the place in the error chain.
- `contract` and `contract.go` moved into the `contract` package; the root
  `contract.go` was removed.
- Every command is registered through `adapt`, which centralises printing the
  result and choosing the exit code.
- `tycl gen` accepts `tycl` as an output format, not only `json|yaml|toml`.

### Fixed

- `tycl set` rewrites the file in canonical form and can change a key's type
  without leaving a stale duplicate behind.
- Unknown paths in `get`/`set`/`query` are rejected with a "did you mean" hint
  instead of an empty value or a panic.
- `docs` and `types` no longer pollute stderr during normal operation.

### Migration

- `lc` v1 -> v2: rewrite the imports `github.com/pt-main/lc/...` to
  `github.com/pt-main/lc/v2/...`, then run
  `go get github.com/pt-main/lc/v2@v2.0.1`. A reference migration script is
  available in the Lc repository: `docs/to-2.0-migrate.sh`.
- `tap` -> `tap/go`: rewrite `github.com/pt-main/tap/...` to
  `github.com/pt-main/tap/go/...`. The package name stays `tap`, so aliases
  need no changes.
- `tycl file <subcommand> --path=<file> ...` -> `tycl <subcommand> <file> ...`:

  | Was                                         | Now                        |
  |---------------------------------------------|----------------------------|
  | `tycl file get --path=c.tycl int port`      | `tycl get c.tycl port`     |
  | `tycl file set --path=c.tycl int port 9090` | `tycl set c.tycl port auto 9090` |
  | `tycl file remove --path=c.tycl int port`   | `tycl remove c.tycl port`  |
  | `tycl file structure --path=c.tycl`         | `tycl structure c.tycl`    |

---

## Change types

- **Added** - new functionality.
- **Changed** - changes in existing functionality.
- **Deprecated** - soon to be removed functionality.
- **Removed** - removed functionality.
- **Fixed** - bug fixes.
- **Security** - vulnerability fixes.
- **Breaking** - API breaking changes (called out in 1.4.0).

## Links

- [GitHub](https://github.com/pt-main/tycl)
- [Go Reference](https://pkg.go.dev/github.com/pt-main/tycl)
