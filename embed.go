//go:build !dev

package kite

import "embed"

//go:embed all:web/admin/dist all:web/template
var AdminFS embed.FS
