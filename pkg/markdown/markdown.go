package markdown

import (
	"bytes"
	"html/template"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type MarkdownParser struct {
	goldmark.Markdown
}

func NewParser() MarkdownParser {
	return MarkdownParser{
		goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Typographer,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
			),
		),
	}
}

func (m *MarkdownParser) ToHTML(path string) (*template.HTML, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var html bytes.Buffer

	err = m.Convert(content, &html)
	if err != nil {
		return nil, err
	}

	rawHTML := template.HTML(html.String())

	return &rawHTML, nil
}