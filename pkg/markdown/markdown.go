package markdown

import (
	"strings"

	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"mvdan.cc/xurls/v2"
)

// Convert a markdown source to HTML.
func Convert(source string) (string, error) {
	transformer := &astTransformer{}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([]string{
					"http:",
					"https:",
				}),
				extension.WithLinkifyURLRegexp(
					xurls.Strict(),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(parser.NewParagraphParser(), 100),
			),
			parser.WithInlineParsers(
				util.Prioritized(parser.NewLinkParser(), 100),
				util.Prioritized(parser.NewAutoLinkParser(), 200),
				util.Prioritized(parser.NewEmphasisParser(), 300),
			),
			parser.WithASTTransformers(
				util.Prioritized(transformer, 100),
			),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	var buf strings.Builder
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}

	if transformer.err != nil {
		return "", transformer.err
	}

	return buf.String(), nil
}

type astTransformer struct {
	err error
}

func (t *astTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	t.err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Link:
			v.SetAttributeString("target", "_blank")
			v.SetAttributeString("rel", "noopener")

		case *ast.AutoLink:
			v.SetAttributeString("target", "_blank")
			v.SetAttributeString("rel", "noopener")

		case *ast.Document:
		case *ast.Paragraph:
		case *ast.Emphasis:
		case *ast.Text:
		case *east.Strikethrough:
		case *ast.RawHTML:
		case *ast.HTMLBlock:

		default:
			return 0, wreck.InvalidValue.New("Invalid markdown")
		}

		return ast.WalkContinue, nil
	})
}
