---
PLAN: "feat: path helpers moved out of webtyp/fmt, plus Tilde"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — filepath: the path helpers leave `fmt`, and gain `Tilde`

Phase **A1 (gate)** of the master plan
`SOURCE_SELECTION_MASTER_PLAN.md` (orchestration only — everything this plan needs is inline).
`webtyp/devbrowser`, `webtyp/app`, `webtyp/devtui`, `webtyp/fetch`, `webtyp/ddlc` migrate to this
repo, and only then does `webtyp/fmt` delete its copy.

Read [AGENTS.md](../AGENTS.md) first. Critical rules, repeated:
- This package compiles to **TinyGo/WASM**. Do NOT import `strings`, `strconv`, `errors`, stdlib
  `fmt` or `path/filepath`. Use `webtyp.com/fmt`'s exported `Index`, `HasSuffix`, `HasPrefix`.
- `os` only in `*.stlib.go` (`//go:build !wasm`); `syscall/js` only in `*.wasm.go` (`//go:build wasm`).
- No `map`, no `reflect`. Tests never depend on the real home or the process CWD.

## Why

`webtyp.com/fmt` keeps growing, and it is being slimmed down to the minimum every WASM binary needs.
Path handling is a separate concern: most binaries never use it. It moves here **with its tests**.

It also gains `Tilde`: write the home directory as `~` in text shown to people and LLMs. That
replaces `webtyp/app`'s private `AbbreviateHome`, which only handled whole strings, not paths
embedded in log lines.

## State when this plan starts (moved by the maintainer, 2026-10-06)

The files were physically moved from `webtyp/fmt` **unchanged**. They still say `package fmt`, so
nothing here compiles yet. Fixing that is this plan's job.

| File here | Was in `webtyp/fmt` |
|---|---|
| `filepath.go`, `filepath.stlib.go`, `filepath.wasm.go` | same names |
| `tests/filepath_test.go` | `filepath_test.go` |
| `tests/short_test.go` | `filepath.short_test.go` |
| `docs/API_FILEPATH.md` | `docs/API_FILEPATH.md` |

The gonew stub (`type Filepath struct{}` / `New()`) was deleted. Do not recreate it.

RULE for this repo: **every test lives in `tests/`**, as `package filepath_test`, importing
`webtyp.com/filepath` and using only the exported API. **Never export a symbol only so a test can
reach it.** A test that truly needs an unexported identifier stays at the root (`package filepath`),
with a comment at the top of the file explaining what private thing it needs and why the public API
cannot observe it. This plan needs none.

## Design gate

**1. Prior art.**
- Go `path/filepath`: `Join`, `Base`, `Ext`, `Rel`. The names a Go developer already knows.
- Node `path` (`join`, `basename`, `extname`, `relative`).
- Python `os.path` (`join`, `basename`, `splitext`, `relpath`, `expanduser`).

We keep the stdlib's names where the meaning is identical (`Join`, `Base`, `Ext`). We do NOT reuse
`Rel`: the stdlib's `Rel` returns `(string, error)` and handles one path. Ours rewrites every
occurrence of a base inside a text into `./`, so it keeps its own name, `RelativeTo`. Why this
package exists at all: the stdlib's `path/filepath` drags `os`/`syscall` into a TinyGo WASM binary.

**2. Novice-name test.** The old method-on-`Conv` API cannot move: Go forbids declaring methods
on `fmt.Conv` outside package `fmt`, and the code used `Conv`'s unexported buffers. Plain functions
replace it.

| Old (`webtyp.com/fmt`) | New (`webtyp.com/filepath`) | Read aloud |
|---|---|---|
| `PathJoin(a, b).String()` | `Join(a, b)` | join the parts |
| `Convert(p).PathBase().String()` | `Base(p)` | the last element |
| `Convert(p).PathExt().String()` | `Ext(p)` | the extension |
| `Convert(s).PathShort().String()` | `Short(s)` | shorten paths under the root to `./` |
| `PathRelativeTo(s, base)` | `RelativeTo(s, base)` | shorten paths under base to `./` |
| `SetPathBase(dir)` | deleted | its only user outside fmt was a test (devtui); a caller with its own base uses `RelativeTo` |
| `GetPathBase()` | unexported `detectRoot()` | no consumer outside fmt used it |
| — | `Tilde(s)` | write the home directory as `~` |

No setter exists for the root or the home. Tests control them with `t.Chdir` and
`t.Setenv("HOME", …)`, which are standard test tools, instead of with API added for them.

**3. Complexity ledger.**
```
Concepts the developer must learn   +1 (Tilde) / −3 (no Conv chaining for paths, no SetPathBase, no GetPathBase)
Files they must touch to do X       +0 / −0
Lines at the call site              −1 per call (no .String())
Ways to do the same thing           +0 / −2  (fmt's path API and app.AbbreviateHome die in later phases)
```

**4. Where it belongs.** A path concern, separate from formatting. One repo, used by every
consumer.

**5. What it deletes.** Here: nothing, because this is the destination. Later phases delete
`fmt/filepath*.go`, `fmt/docs/API_FILEPATH.md` and `app.AbbreviateHome`.

## Stage 1 — move and convert (no behaviour change)

1. `package fmt` → `package filepath` in the three source files. `go get webtyp.com/fmt@latest` for
   `Index`/`HasSuffix`/`HasPrefix`.
2. Rewrite the implementation without `Conv`. Every function takes and returns `string`. Build
   results in a local `[]byte` (`buf = append(buf, s...)`) and return `string(buf)`.
   - `PathJoin` → `func Join(elem ...string) string`, same algorithm and outputs.
   - `PathBase` → `func Base(path string) string`; `PathExt` → `func Ext(path string) string`.
     Keep `pathClean` and `extractBase` as unexported helpers, unchanged.
   - `shortenAgainst` → `func shortenAgainst(text, base, replacement string) string`. Same matching
     loop, written against `text` instead of `BuffOut`. Where it wrote `"."` / `"./"` it writes
     `replacement` / `replacement + "/"`. `Short` and `RelativeTo` pass `"."`.
   - Delete `var pathBase` and `SetPathBase`. There is no package-level state.
   - `PathShort` → `func Short(text string) string` = `RelativeTo(text, detectRoot())`. It
     detects on **every call**, with no cache, so the result follows the current directory.
     `os.Getwd` is cheap.
   - `PathRelativeTo` → `func RelativeTo(text, base string) string`.
   - `GetPathBase` in `filepath.stlib.go`/`filepath.wasm.go` → unexported `detectRoot()`, same
     bodies.
3. Convert the tests in `tests/` to `package filepath_test`. Map them mechanically:
   `Convert(x).PathBase().String()` → `filepath.Base(x)`, and so on. The `TestPathShort` cases set
   a base through the private `pathBase`. Rewrite each as a `RelativeTo(input, base)` case: same
   algorithm, explicit base, works under WASM too. Keep exactly two cases for `Short` itself, in
   a separate file `tests/short_cwd_test.go` with `//go:build !wasm`: `t.Chdir(dir)` with
   `dir := t.TempDir()`, then `filepath.Short(dir + "/a.go") == "./a.go"`, and a path outside
   `dir` stays unchanged.
   `TestPathExtNormalizeCase` / `TestPathJoinNormalizeCase` chain `ToLower`: write them as
   `fmt.Convert(filepath.Ext(x)).ToLower().String()`. Every existing case keeps its exact expected
   output.
4. `gotest` green **before** Stage 2: this stage is a pure move.

## Stage 2 — boundary fix, red test first

A non-root base currently matches in the middle of another path. With root `/home/user/project`,
`see /mnt/home/user/project/x` becomes `see /mnt./x`.

1. Add to the `RelativeTo` cases (in `tests/short_test.go`) the case
   `"base inside another path is not a match"`: base `/home/user/project`, input
   `see /mnt/home/user/project/x`, want unchanged. Run it; it fails.
2. Fix: a match is valid only if it starts at a path boundary — `matchIdx == 0`, or the previous byte
   is one of `' ' '\t' '\n' '\r' '"' '\'' '(' '`' '='`. Put this in ONE helper
   `isPathStart(prev byte) bool`, and use it in the root (`/`) branch too. Its current set is a
   subset; adding `` ` `` and `=` there is intended. All existing cases stay green.

## Stage 3 — `Tilde` (red tests first, new test file `tests/tilde_test.go`)

API (in `filepath.go`):

```go
// Tilde writes every occurrence of the user's home directory in text as "~"
// ("/home/dev/Project/app" → "~/Project/app"), including paths embedded in log
// lines, like Short. No-op when the home is unknown (always in WASM) or is "/".
func Tilde(text string) string
```

`filepath.stlib.go`: `func detectHome() string` returns `os.UserHomeDir()` cleaned with
`pathClean`, or `""` on error. `filepath.wasm.go`: `func detectHome() string { return "" }`.
`Tilde`: `home := detectHome()` on **every call** (no cache, so `HOME` changes are seen;
`os.UserHomeDir` only reads an env var). If `home` is `""`, `"/"` or `"\"`, return `text`.
Otherwise return `shortenAgainst(text, home, "~")`.

Tests go in `tests/tilde_test.go` with `//go:build !wasm` (in WASM there is no home, and `Tilde` is
a no-op by design). Each test starts with `t.Setenv("HOME", "/home/dev")`. On Unix,
`os.UserHomeDir` reads `$HOME`. `t.Setenv` restores the value afterwards, and the real home is never
touched.

| name | in | want |
|---|---|---|
| under home | `/home/dev/Project/app` | `~/Project/app` |
| home itself | `/home/dev` | `~` |
| sibling with same prefix | `/home/developer/app` | unchanged |
| outside home | `/etc/hosts` | unchanged |
| relative | `web/server.go` | unchanged |
| embedded in log | `open /home/dev/a.go and /home/dev/b.go` | `open ~/a.go and ~/b.go` |
| in backticks | ``file `/home/dev/x.go` `` | ``file `~/x.go` `` |
| inside another path | `/mnt/home/dev/x` | unchanged |

Plus `TestTildeRootHome`: `t.Setenv("HOME", "/")` → `/etc/hosts` unchanged. And
`TestTildeNoHome`: `t.Setenv("HOME", "")` → input unchanged (`os.UserHomeDir` errors).

## Stage 4 — README

Write `README.md` from `docs/API_FILEPATH.md` (then delete that file), rewritten for the new names. Start with an
"I want X → use Y" table (Join, Base, Ext, Short, RelativeTo, Tilde). Then add the migration table
from the design gate under `## Migrating from webtyp.com/fmt`, and the note on the `wpath` import
alias for backend code that also imports the stdlib's `path/filepath`.

## Acceptance

- `gotest` passes (includes WASM).
- `grep -rn '"strings"\|"strconv"\|"errors"\|"path/filepath"' --include='*.go' .` → empty.
- `grep -rln '"os"' --include='*.go' .` → only `filepath.stlib.go` and files under `tests/`.
- `ls *_test.go 2>/dev/null` → nothing at the root.
- `grep -rn "Conv\b\|\.String()" filepath.go` → no use of `fmt.Conv`.
- `grep -rn "func Set\|^var " --include='*.go' .` → empty (no setters, no package state).

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Convert | `filepath.go`, `filepath.stlib.go`, `filepath.wasm.go`, `tests/filepath_test.go`, `tests/short_test.go`, `tests/short_cwd_test.go`, `go.mod`, `go.sum` |
| 2 | Boundary fix | `filepath.go`, `tests/short_test.go` |
| 3 | Tilde | `filepath.go`, `filepath.stlib.go`, `filepath.wasm.go`, `tests/tilde_test.go` |
| 4 | README | `README.md`, `docs/API_FILEPATH.md` (deleted) |
