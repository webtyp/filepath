package filepath_test

import (
	"testing"
	"webtyp.com/filepath"
)

// TestPathRelativeTo proves shortenAgainst behaves identically whether reached
// via Short's cached pathBase global or via RelativeTo's explicit
// base parameter.
func TestPathRelativeTo(t *testing.T) {
	tests := []struct {
		name string
		base string
		path string
		want string
	}{
		{
			name: "relative from manual base",
			base: "/home/user/project",
			path: "/home/user/project/modules/test.js",
			want: "./modules/test.js",
		},
		{
			name: "exactly same as base",
			base: "/home/user/project",
			path: "/home/user/project",
			want: ".",
		},
		{
			name: "different path",
			base: "/home/user/project",
			path: "/etc/passwd",
			want: "/etc/passwd",
		},
		{
			name: "prefix but not subpath",
			base: "/home/user/pro",
			path: "/home/user/project",
			want: "/home/user/project",
		},
		{
			name: "subpath with trailing slash in input",
			base: "/home/user/project",
			path: "/home/user/project/web/",
			want: "./web/",
		},
		{
			name: "manually set base as root",
			base: "/",
			path: "/etc/passwd",
			want: "./etc/passwd",
		},
		{
			name: "embedded path in log message",
			base: "/home/user/Dev/Pkg/webtyp/app/example",
			path: "Compiling WASM due to /home/user/Dev/Pkg/webtyp/app/example/web/client.go change... ",
			want: "Compiling WASM due to ./web/client.go change... ",
		},
		{
			name: "another embedded path in log message",
			base: "/home/user/Dev/Pkg/webtyp/app/example",
			path: " 13:07:52  ASSETS  .js create ... /home/user/Dev/Pkg/webtyp/app/example/modules/users/newfile.js",
			want: " 13:07:52  ASSETS  .js create ... ./modules/users/newfile.js",
		},
		{
			name: "path at the end of sentence",
			base: "/home/user/Dev/Pkg/webtyp/app/example",
			path: "WASM source file already exists at /home/user/Dev/Pkg/webtyp/app/example/web/client.go , skipping generation",
			want: "WASM source file already exists at ./web/client.go , skipping generation",
		},
		{
			name: "multiple occurrences",
			base: "/home/user/project",
			path: "moving /home/user/project/a to /home/user/project/b",
			want: "moving ./a to ./b",
		},
		{
			name: "base inside another path is not a match",
			base: "/home/user/project",
			path: "see /mnt/home/user/project/x",
			want: "see /mnt/home/user/project/x",
		},
		{
			name: "within quotes",
			base: "/home/user/project",
			path: `source is "/home/user/project/main.go"`,
			want: `source is "./main.go"`,
		},
		{
			name: "base followed by punctuation",
			base: "/home/user/project",
			path: "current dir is /home/user/project, check it.",
			want: "current dir is /home/user/project, check it.", // NOT valid boundary because , is not / or \
		},
		{
			// The case that matters for webtyp/ddlc's daemon-side Label():
			// a daemon whose own CWD doesn't track the project root can still
			// shorten its output path for display using its known root.
			name: "ddlc export path relative to project root",
			base: "/home/user/project",
			path: "/home/user/project/config/db.sql",
			want: "./config/db.sql",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filepath.RelativeTo(tc.path, tc.base)
			if got != tc.want {
				t.Errorf("RelativeTo(%q, %q) = %q; want %q", tc.path, tc.base, got, tc.want)
			}
		})
	}
}

func TestPathShortWindows(t *testing.T) {
	// Manual test for windows-style paths even on linux
	// since pathClean and Join handle them conceptually
	base := `C:\Users\Project`

	got := filepath.RelativeTo(`C:\Users\Project\file.txt`, base)
	want := "./file.txt"
	if got != want {
		t.Errorf("Windows relative: got %q; want %q", got, want)
	}

	got = filepath.RelativeTo(`C:\Users\Project`, base)
	want = "."
	if got != want {
		t.Errorf("Windows same: got %q; want %q", got, want)
	}
}
