// Package loginassets exposes the login visuals to the embedded fallback page.
package loginassets

import "embed"

// Assets contains the visuals shared by the Vue login page and the backend
// fallback page.
//
//go:embed bg.png gops.svg illustration.svg
var Assets embed.FS
