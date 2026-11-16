package filepath

import (
	"webtyp.com/fmt"
)

// CompactList renders paths as one string: the common parent once (with ~ when
// under home) and only the differing last element in braces.
//   CompactList([]string{"/home/u/P/webtyp/form", "/home/u/P/webtyp/model"})
//   -> "~/P/webtyp/{form, model}"
// One path -> Tilde(path). Zero -> "". Paths with no common parent beyond "/"
// are joined with ", " each through Tilde. Input order is preserved.
func CompactList(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return Tilde(paths[0])
	}

	cleaned := make([]string, len(paths))
	for i, p := range paths {
		c, _ := pathClean(p)
		cleaned[i] = c
	}

	sep := byte('/')
	for _, p := range cleaned {
		if fmt.Index(p, "\\") != -1 {
			sep = '\\'
			break
		}
	}

	common := findCommonDirPrefix(cleaned, sep)
	if !isBeyondRoot(common) {
		var buf []byte
		for i, p := range cleaned {
			if i > 0 {
				buf = append(buf, ", "...)
			}
			buf = append(buf, Tilde(p)...)
		}
		return string(buf)
	}

	prefixLen := len(common)
	var buf []byte
	buf = append(buf, common...)
	buf = append(buf, '{')
	for i, p := range cleaned {
		if i > 0 {
			buf = append(buf, ", "...)
		}
		rem := p[prefixLen:]
		buf = append(buf, rem...)
	}
	buf = append(buf, '}')

	return Tilde(string(buf))
}

func isBeyondRoot(common string) bool {
	if common == "" || common == "." {
		return false
	}
	if common == "/" || common == "\\" {
		return false
	}
	if len(common) == 3 && common[1] == ':' && (common[2] == '\\' || common[2] == '/') {
		return false
	}
	return true
}

func findCommonDirPrefix(paths []string, sep byte) string {
	if len(paths) == 0 {
		return ""
	}

	prefixes := getDirPrefixes(paths[0], sep)
	for i := len(prefixes) - 1; i >= 0; i-- {
		pref := prefixes[i]
		matchAll := true
		for j := 1; j < len(paths); j++ {
			if !fmt.HasPrefix(paths[j], pref) {
				matchAll = false
				break
			}
		}
		if matchAll {
			return pref
		}
	}
	return ""
}

func getDirPrefixes(p string, sep byte) []string {
	var prefixes []string

	if len(p) > 0 && p[0] == sep {
		prefixes = append(prefixes, string([]byte{sep}))
	} else if len(p) >= 3 && p[1] == ':' && p[2] == sep {
		prefixes = append(prefixes, p[:3])
	}

	for i := 0; i < len(p); i++ {
		if p[i] == sep {
			pref := p[:i+1]
			if len(prefixes) == 0 || prefixes[len(prefixes)-1] != pref {
				prefixes = append(prefixes, pref)
			}
		}
	}
	return prefixes
}
