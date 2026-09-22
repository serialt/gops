// Package webui contains the frontend files embedded in the backend binary.
package webui

import "embed"

// Assets contains the production frontend build. The Makefile replaces the
// fallback page in dist while compiling and restores it afterwards.
//
//go:embed all:dist
var Assets embed.FS
