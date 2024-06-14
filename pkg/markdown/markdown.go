package markdown

import (
	"bytes"
	"html/template"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type MarkdownParser struct {
	goldmark.Markdown
}

type ASTTransformer struct{}

func (a *ASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Link:
			if string(v.Destination)[0] != '/' {
				v.SetAttribute([]byte("target"), "_blank")
			}
		}

		return ast.WalkContinue, nil
	})
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
				parser.WithASTTransformers(util.Prioritized(&ASTTransformer{}, 42)),
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
