# webtyp.com/filepath
<img src="docs/img/badges.svg">

Path helpers for the webtyp ecosystem that compile to TinyGo/WASM.

This package provides path manipulation helpers (`Join`, `Base`, `Ext`) and display helpers that rewrite paths inside text for people and LLMs (`Short`, `RelativeTo`, `Tilde`). It acts as a WASM-friendly replacement for the Go standard library's `path/filepath`.

## I want X -> use Y

| I want... | use... |
|-----------|--------|
| Join path elements | `filepath.Join("a", "b", "c")` |
| Get the last element of a path | `filepath.Base(path)` |
| Get the file extension | `filepath.Ext(path)` |
| Shorten absolute paths relative to CWD | `filepath.Short(text)` |
| Shorten paths relative to a custom base | `filepath.RelativeTo(text, base)` |
| Replace home directory with `~` | `filepath.Tilde(text)` |

### Examples

#### `Join`
```go
filepath.Join("a", "b", "c")            // -> "a/b/c"
filepath.Join("/root", "sub", "file")   // -> "/root/sub/file"
```

#### `Base`
```go
filepath.Base("/a/b/c.txt") // -> "c.txt"
filepath.Base("folder/file.txt")   // -> "file.txt"
```

#### `Ext`
```go
filepath.Ext("file.txt")          // -> ".txt"
filepath.Ext("/path/to/archive.tar.gz") // -> ".gz"
```

#### `Short`
Shortens paths to a `./` relative form, auto-detecting the CWD. Works on embedded paths in text.
```go
filepath.Short("Compiling /home/user/project/web/client.go")
// -> "Compiling ./web/client.go"
```

#### `RelativeTo`
Like `Short`, but with an explicit base rather than using auto-detection.
```go
filepath.RelativeTo("/home/user/project/config/db.sql", "/home/user/project")
// -> "./config/db.sql"
```

#### `Tilde`
Replaces the user's home directory with `~`. Works on embedded paths.
```go
filepath.Tilde("open /home/dev/a.go and /home/dev/b.go")
// -> "open ~/a.go and ~/b.go"
```

---

## Migrating from webtyp.com/fmt

The path manipulation functions have been moved out of `fmt` into this package, with some simplifications:

| Old (`webtyp.com/fmt`) | New (`webtyp.com/filepath`) |
|---|---|
| `PathJoin(a, b).String()` | `Join(a, b)` |
| `Convert(p).PathBase().String()` | `Base(p)` |
| `Convert(p).PathExt().String()` | `Ext(p)` |
| `Convert(s).PathShort().String()` | `Short(s)` |
| `PathRelativeTo(s, base)` | `RelativeTo(s, base)` |
| `SetPathBase(dir)` | *deleted* |
| `GetPathBase()` | *unexported `detectRoot()`* |
| — | `Tilde(s)` |

> **Note on import aliasing**: Since this package uses the name `filepath`, backend code that also imports the standard library's `path/filepath` should import this package with an alias, typically `wpath "webtyp.com/filepath"`.
