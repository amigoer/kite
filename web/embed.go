package web

import "embed"

//go:embed all:admin/dist all:template
var AdminFS embed.FS
