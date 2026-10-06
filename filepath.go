package filepath

import (
	"webtyp.com/fmt"
)

// Join joins path elements using the appropriate separator.
// Accepts variadic string arguments and returns a string.
// Detects Windows paths (backslash) or Unix paths (forward slash).
// Empty elements are ignored.
//
// Usage patterns:
//   - Join("a", "b", "c")           // -> "a/b/c"
//
// Examples:
//
//	Join("a", "b", "c")           // -> "a/b/c"
//	Join("/root", "sub", "file")   // -> "/root/sub/file"
//	Join(`C:\dir`, "file")        // -> "C:\dir\file"
//	Join("a", "", "b")            // -> "a/b"
func Join(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}

	sep := "/"
	// detect separator from first element with a separator
	for _, e := range elem {
		if fmt.Index(e, "\\") != -1 {
			sep = "\\"
			break
		}
	}

	var buf []byte

	for i, e := range elem {
		if e == "" {
			continue
		}

		// trim leading separators only if not the first element
		if i > 0 && len(buf) > 0 {
			for len(e) > 0 && (e[0] == '/' || e[0] == '\\') {
				e = e[1:]
			}
		}

		// add separator if needed
		if len(buf) > 0 && !fmt.HasSuffix(string(buf), sep) && e != "" {
			buf = append(buf, sep...)
		}
		buf = append(buf, e...)
	}

	return string(buf)
}

// pathClean normalizes a path by detecting the separator and handling special cases.
// Returns the cleaned path and the detected separator.
// This is a helper function used by Base and Ext to avoid code duplication.
func pathClean(path string) (string, byte) {
	if path == "" {
		return ".", '/'
	}

	// prefer backslash if present
	sep := byte('/')
	if fmt.Index(path, "\\") != -1 {
		sep = '\\'
	}

	// windows drive root like "C:\" or with only extra separators -> return "\\"
	if sep == '\\' && len(path) >= 2 && path[1] == ':' {
		onlySep := true
		for i := 2; i < len(path); i++ {
			if path[i] != '\\' && path[i] != '/' {
				onlySep = false
				break
			}
		}
		if onlySep {
			return "\\", sep
		}
	}

	// trim trailing separators
	for len(path) > 1 && path[len(path)-1] == sep {
		path = path[:len(path)-1]
	}

	// if path reduced to a single root separator, return it
	if len(path) == 1 && (path[0] == '/' || path[0] == '\\') {
		return path, sep
	}

	return path, sep
}

// extractBase returns the base filename from a cleaned path if prefix is empty.
// If prefix is set, it attempts to return the relative path from that prefix.
func extractBase(cleaned string, sep byte, prefix string) string {
	// If prefix is set, try to strip it
	if prefix != "" {
		if fmt.HasPrefix(cleaned, prefix) {
			rel := cleaned[len(prefix):]
			// check if it's a full component match
			isRoot := len(prefix) == 1 && (prefix[0] == '/' || prefix[0] == '\\')
			if isRoot || len(rel) == 0 || rel[0] == '/' || rel[0] == '\\' {
				if len(rel) > 0 && (rel[0] == '/' || rel[0] == '\\') {
					return rel[1:]
				}
				return rel
			}
		}
		return cleaned
	}

	// Default behavior: extract filename after last separator
	// Special cases
	if cleaned == "." || cleaned == "\\" || cleaned == "/" {
		return ""
	}

	// search from end for last separator
	for i := len(cleaned) - 1; i >= 0; i-- {
		if cleaned[i] == sep {
			return cleaned[i+1:]
		}
	}
	// no separator found - whole cleaned path is the base
	return cleaned
}

// Base returns the last element of path, similar to
// filepath.Base from the Go standard library. It treats
// trailing slashes specially ("/a/b/" -> "b") and preserves
// a single root slash ("/" -> "/"). An empty path returns ".".
//
// Examples:
//
//		Base("/a/b/c.txt") // -> "c.txt"
//		Base("folder/file.txt")   // -> "file.txt"
//		Base("")           // -> "."
//	 Base(`c:\file program\app.exe`) // -> "app.exe"
func Base(path string) string {
	cleaned, sep := pathClean(path)

	base := extractBase(cleaned, sep, "")
	if base == "" {
		// Special case: write the cleaned value (., /, or \)
		return cleaned
	}
	return base
}

// Ext extracts the file extension from a path.
// An empty extension returns an empty string.
//
// Examples:
//
//	Ext("/a/b/c.txt") // -> ".txt"
//	Ext("file.tar.gz") // -> ".gz"
//	Ext("noext")       // -> ""
func Ext(path string) string {
	cleaned, sep := pathClean(path)

	// get the base filename using helper
	base := extractBase(cleaned, sep, "")
	if base == "" {
		// Special cases like ".", "/", "\\" have no extension
		return ""
	}

	// special cases: "." and ".." have no extension
	if base == "." || base == ".." {
		return ""
	}

	// search for last dot in base filename
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			// don't count leading dot (hidden files like .bashrc)
			if i == 0 {
				return ""
			}
			return base[i:]
		}
	}

	return ""
}

// Short shortens absolute paths relative to root path.
// It can handle paths embedded in larger strings (e.g. log messages).
// Auto-detects root path on every call.
// Returns relative path with "./" prefix for minimal output.
// Example: "Compiling /home/user/project/src/file.go ..." -> "Compiling ./src/file.go ..."
func Short(text string) string {
	root := detectRoot()
	if root == "" {
		return text
	}

	return RelativeTo(text, root)
}

// RelativeTo shortens text's occurrences of base into "./"-relative form,
// using the same algorithm as Short but with the base given explicitly —
// for callers that track their own reference directory instead of relying on
// CWD auto-detection.
//
// Example: RelativeTo("/home/user/project/config/db.sql", "/home/user/project")
// -> "./config/db.sql"
func RelativeTo(text, base string) string {
	cleanedBase, _ := pathClean(base)
	if cleanedBase == "" {
		return text
	}

	return shortenAgainst(text, cleanedBase, ".")
}

func isPathStart(prev byte) bool {
	return prev == ' ' || prev == '\t' || prev == '\n' || prev == '\r' || prev == '"' || prev == '\'' || prev == '(' || prev == '`' || prev == '='
}

// Tilde writes every occurrence of the user's home directory in text as "~"
// ("/home/dev/Project/app" -> "~/Project/app"), including paths embedded in log
// lines, like Short. No-op when the home is unknown (always in WASM) or is "/".
func Tilde(text string) string {
	home := detectHome()
	if home == "" || home == "/" || home == "\\" {
		return text
	}
	return shortenAgainst(text, home, "~")
}

// shortenAgainst rewrites occurrences of base in text into replacement-relative form.
// base must already be cleaned (see pathClean).
func shortenAgainst(text, base, replacement string) string {
	if text == "" {
		return ""
	}

	var buf []byte
	start := 0

	for {
		idx := fmt.Index(text[start:], base)
		if idx == -1 {
			buf = append(buf, text[start:]...)
			break
		}

		matchIdx := start + idx
		buf = append(buf, text[start:matchIdx]...)

		// Validate match boundary
		endIdx := matchIdx + len(base)
		isRoot := len(base) == 1 && (base[0] == '/' || base[0] == '\\')

		valid := false

		if matchIdx == 0 || isPathStart(text[matchIdx-1]) {
			if isRoot {
				valid = true
				// Root followed by another separator is not a valid single root match (e.g. //)
				if valid && endIdx < len(text) && (text[endIdx] == '/' || text[endIdx] == '\\') {
					valid = false
				}
			} else {
				if endIdx == len(text) {
					valid = true
				} else {
					nextChar := text[endIdx]
					if nextChar == '/' || nextChar == '\\' {
						valid = true
					}
				}
			}
		}

		if valid {
			if isRoot {
				if endIdx == len(text) {
					buf = append(buf, replacement...)
				} else {
					buf = append(buf, replacement...)
					buf = append(buf, '/')
				}
				start = endIdx
			} else {
				buf = append(buf, replacement...)

				// If followed by a separator, consume it and write "/" to normalize
				if endIdx < len(text) && (text[endIdx] == '/' || text[endIdx] == '\\') {
					buf = append(buf, '/')
					start = endIdx + 1
				} else {
					start = endIdx
				}
			}
		} else {
			// Not a valid path boundary, just copy the match and continue
			buf = append(buf, base...)
			start = endIdx
		}
	}

	return string(buf)
}
