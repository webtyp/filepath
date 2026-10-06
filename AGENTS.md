# AGENTS.md — webtyp/filepath

Working notes for AI agents operating in this repository. End-user docs: [README.md](README.md).

## What this repo is

Path helpers for the webtyp ecosystem that compile to TinyGo/WASM: `Join`, `Base`, `Ext`, and the
display helpers that rewrite paths inside text for people and LLMs — `Short` (project root → `./`),
`RelativeTo` (explicit base → `./`) and `Tilde` (home → `~`). It was split out of `webtyp.com/fmt`
so `fmt` stays small.

## This package compiles to WASM. No standard library except behind build tags.

- Do NOT import `strings`, `strconv`, `errors`, stdlib `fmt` or `path/filepath`. Use
  `webtyp.com/fmt` exported helpers (`fmt.Index`, `fmt.HasSuffix`, `fmt.HasPrefix`) — never copy
  them here.
- `os` and `syscall/js` only in build-tagged files: `*.stlib.go` (`//go:build !wasm`) and
  `*.wasm.go` (`//go:build wasm`), one function per environment behind the same name.
- No `map`. No `reflect`.
- The package name equals the stdlib's `path/filepath` on purpose (like `webtyp.com/fmt` and
  `fmt`): it is the WASM replacement. Backend code that also uses the stdlib imports this one as
  `wpath "webtyp.com/filepath"`.

## The build that defines "done"

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # once
gotest            # vet + race + cover + wasm
```

## Rules

- Every test lives in `tests/` (`package filepath_test`, public API only). A root-level test is
  allowed only with a comment justifying the unexported identifier it needs. **Never export a
  symbol just so a test can reach it.**
- Tests control the environment with test tools, not API: `t.Chdir` for `Short`,
  `t.Setenv("HOME", …)` for `Tilde`, `//go:build !wasm` where the behaviour only exists on the
  backend. Prefer `RelativeTo` (explicit base) to test the shortening algorithm.
- No package-level state: `Short` and `Tilde` detect the root/home on every call.
- Every function works on both `/` and `\` separators (the separator is detected from the input),
  as the original `fmt` implementation did.
