package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/frontmatter"
)

type CVStore struct {
	byLocale map[string]string
}

func LoadCV(fsys fs.FS) (*CVStore, error) {
	store := &CVStore{byLocale: make(map[string]string)}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("reading CV directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		locale := extractCVLocale(entry.Name())
		if !isSupportedLocale(locale) {
			continue
		}

		src, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading CV %s: %w", entry.Name(), err)
		}

		htmlContent, err := convertCV(src)
		if err != nil {
			return nil, fmt.Errorf("converting CV %s: %w", entry.Name(), err)
		}

		store.byLocale[locale] = htmlContent
	}

	return store, nil
}

func (s *CVStore) Get(locale string) string {
	if html, ok := s.byLocale[locale]; ok {
		return html
	}
	if html, ok := s.byLocale[defaultLocale]; ok {
		return html
	}
	return ""
}

func extractCVLocale(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	if idx := strings.LastIndex(name, "."); idx > 0 {
		return name[idx+1:]
	}
	return defaultLocale
}

func convertCV(src []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			&frontmatter.Extender{
				Mode: frontmatter.SetMetadata,
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
			parser.WithASTTransformers(util.Prioritized(cvClassTransformer{}, 100)),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type cvClassTransformer struct{}

func (cvClassTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Heading:
			node.SetAttributeString("class", []byte(headingClass(node.Level)))
		case *ast.Paragraph:
			if _, ok := n.Parent().(*ast.ListItem); ok {
				return ast.WalkContinue, nil
			}
			node.SetAttributeString("class", []byte("my-2 print:my-1"))
		case *ast.List:
			node.SetAttributeString("class", []byte(listClass(node.IsOrdered())))
		case *ast.ListItem:
			node.SetAttributeString("class", []byte("leading-snug print:leading-tight"))
		case *ast.ThematicBreak:
			node.SetAttributeString("class", []byte("my-4 border-border/40 print:my-1 print:border-black/30"))
		}
		return ast.WalkContinue, nil
	})
}

func headingClass(level int) string {
	switch level {
	case 1:
		return "text-3xl sm:text-5xl font-bold tracking-wider text-center mb-1 pb-1 border-b border-border print:text-3xl print:mb-1 print:pb-1 print:border-black"
	case 2:
		return "text-lg font-bold mt-4 mb-1 pb-1 border-b border-border print:text-base print:mt-3 print:mb-0.5 print:pb-0.5 print:border-black"
	case 3:
		return "text-base font-bold mt-3 mb-1 print:text-sm print:mt-1.5 print:mb-0"
	default:
		return "font-bold mt-2 mb-1 print:mt-1 print:mb-0"
	}
}

func listClass(ordered bool) string {
	base := "pl-5 space-y-1 my-2 print:my-1 print:space-y-0.5 print:pl-4"
	if ordered {
		return "list-decimal " + base
	}
	return "list-disc " + base
}
