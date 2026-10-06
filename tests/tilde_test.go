//go:build !wasm

package filepath_test

import (
	"testing"
	"webtyp.com/filepath"
)

func TestTilde(t *testing.T) {
	t.Setenv("HOME", "/home/dev")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "under home",
			in:   "/home/dev/Project/app",
			want: "~/Project/app",
		},
		{
			name: "home itself",
			in:   "/home/dev",
			want: "~",
		},
		{
			name: "sibling with same prefix",
			in:   "/home/developer/app",
			want: "/home/developer/app",
		},
		{
			name: "outside home",
			in:   "/etc/hosts",
			want: "/etc/hosts",
		},
		{
			name: "relative",
			in:   "web/server.go",
			want: "web/server.go",
		},
		{
			name: "embedded in log",
			in:   "open /home/dev/a.go and /home/dev/b.go",
			want: "open ~/a.go and ~/b.go",
		},
		{
			name: "in backticks",
			in:   "file `/home/dev/x.go`",
			want: "file `~/x.go`",
		},
		{
			name: "inside another path",
			in:   "/mnt/home/dev/x",
			want: "/mnt/home/dev/x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filepath.Tilde(tc.in)
			if got != tc.want {
				t.Errorf("Tilde(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTildeRootHome(t *testing.T) {
	t.Setenv("HOME", "/")

	in := "/etc/hosts"
	got := filepath.Tilde(in)
	want := "/etc/hosts"

	if got != want {
		t.Errorf("Tilde(%q) with HOME=/ = %q; want %q", in, got, want)
	}
}

func TestTildeNoHome(t *testing.T) {
	t.Setenv("HOME", "")

	in := "/home/dev/x.go"
	got := filepath.Tilde(in)
	want := "/home/dev/x.go"

	if got != want {
		t.Errorf("Tilde(%q) with empty HOME = %q; want %q", in, got, want)
	}
}
