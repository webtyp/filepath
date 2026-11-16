//go:build !wasm

package filepath_test

import (
	"testing"

	"webtyp.com/filepath"
)

func TestCompactList(t *testing.T) {
	t.Setenv("HOME", "/home/cesar")

	tests := []struct {
		name  string
		paths []string
		want  string
	}{
		{
			name:  "vacio",
			paths: []string{},
			want:  "",
		},
		{
			name:  "uno",
			paths: []string{"/home/cesar/Dev/Project/webtyp/form"},
			want:  "~/Dev/Project/webtyp/form",
		},
		{
			name: "dos hermanos",
			paths: []string{
				"/home/cesar/Dev/Project/webtyp/form",
				"/home/cesar/Dev/Project/webtyp/model",
			},
			want: "~/Dev/Project/webtyp/{form, model}",
		},
		{
			name: "cinco hermanos",
			paths: []string{
				"/home/cesar/Dev/Project/webtyp/form",
				"/home/cesar/Dev/Project/webtyp/components",
				"/home/cesar/Dev/Project/webtyp/layout",
				"/home/cesar/Dev/Project/webtyp/widget",
				"/home/cesar/Dev/Project/webtyp/model",
			},
			want: "~/Dev/Project/webtyp/{form, components, layout, widget, model}",
		},
		{
			name: "padres distintos",
			paths: []string{
				"/a/x/f",
				"/a/y/g",
			},
			want: "/a/{x/f, y/g}",
		},
		{
			name: "sin prefijo comun",
			paths: []string{
				"/var/log",
				"/etc/passwd",
			},
			want: "/var/log, /etc/passwd",
		},
		{
			name: "orden preservado",
			paths: []string{
				"/home/cesar/webtyp/zebra",
				"/home/cesar/webtyp/alpha",
				"/home/cesar/webtyp/beta",
			},
			want: "~/webtyp/{zebra, alpha, beta}",
		},
		{
			name: "sin segmentos partidos",
			paths: []string{
				"/a/foo",
				"/a/foobar",
			},
			want: "/a/{foo, foobar}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filepath.CompactList(tc.paths)
			if got != tc.want {
				t.Errorf("CompactList(%v) = %q; want %q", tc.paths, got, tc.want)
			}
		})
	}
}

func TestCompactListNoHome(t *testing.T) {
	t.Setenv("HOME", "")

	paths := []string{
		"/home/cesar/Dev/Project/webtyp/form",
		"/home/cesar/Dev/Project/webtyp/model",
	}
	got := filepath.CompactList(paths)
	want := "/home/cesar/Dev/Project/webtyp/{form, model}"

	if got != want {
		t.Errorf("CompactList without HOME = %q; want %q", got, want)
	}
}
