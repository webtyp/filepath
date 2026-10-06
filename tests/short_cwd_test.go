//go:build !wasm

package filepath_test

import (
	"testing"
	"webtyp.com/filepath"
)

func TestShortCwd(t *testing.T) {
	dir := t.TempDir()

	t.Chdir(dir)

	path := filepath.Join(dir, "a.go")
	got := filepath.Short(path)
	want := "./a.go"

	if got != want {
		t.Errorf("Short(%q) = %q; want %q", path, got, want)
	}

	outsidePath := "/etc/passwd"
	gotOutside := filepath.Short(outsidePath)
	wantOutside := "/etc/passwd"

	if gotOutside != wantOutside {
		t.Errorf("Short(%q) = %q; want %q", outsidePath, gotOutside, wantOutside)
	}
}
