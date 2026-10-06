//go:build wasm

package filepath

import "syscall/js"

// detectRoot returns the domain root path using syscall/js.
func detectRoot() string {
	if global := js.Global(); global.Truthy() {
		if loc := global.Get("location"); loc.Truthy() {
			if origin := loc.Get("origin"); origin.Truthy() {
				return origin.String()
			}
		}
	}
	return "/"
}

// detectHome returns an empty string in WASM as there is no home directory.
func detectHome() string {
	return ""
}
