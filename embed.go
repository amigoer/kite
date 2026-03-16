package kiteblog

import "embed"

//go:embed all:templates ui/admin/dist/*
var resources embed.FS

func TemplateFS() embed.FS {
	return resources
}
