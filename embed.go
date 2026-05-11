package kite

import "embed"

// WebDist holds the compiled web viewer assets.
//
//go:embed all:web/dist
var WebDist embed.FS
