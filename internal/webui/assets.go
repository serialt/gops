// Package webui contains the frontend files embedded in the backend binary.
package webui

import "embed"

// Assets contains the production frontend build and the development fallback.
// The Makefile replaces the placeholder in dist while compiling and restores
// it afterwards.
//
//go:embed fallback.html all:dist
var Assets embed.FS
