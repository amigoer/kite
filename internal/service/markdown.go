package service

import (
	"bytes"
	"fmt"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var postMarkdownRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
	),
)

func renderPostMarkdown(contentMarkdown string) (string, error) {
	var buf bytes.Buffer
	if err := postMarkdownRenderer.Convert([]byte(contentMarkdown), &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("class").OnElements("code", "pre")
	return policy.Sanitize(buf.String()), nil
}
