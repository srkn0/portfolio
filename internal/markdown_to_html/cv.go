package markdowntohtml

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

var cvStore map[string]string

func convertCV(src []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			&frontmatter.Extender{Mode: frontmatter.SetMetadata},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(html.WithXHTML(), html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func LoadCV(fsys fs.FS) error {
	cvStore = make(map[string]string)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading CV directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		locale := "de"
		if idx := strings.LastIndex(name, "."); idx > 0 {
			pl := name[idx+1:]
			if pl == "en" || pl == "de" {
				locale = pl
			}
		}
		src, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return fmt.Errorf("reading CV %s: %w", entry.Name(), err)
		}
		htmlContent, err := convertCV(src)
		if err != nil {
			return fmt.Errorf("converting CV %s: %w", entry.Name(), err)
		}
		cvStore[locale] = htmlContent
	}
	return nil
}

func GetCV(locale string) string {
	if html, ok := cvStore[locale]; ok {
		return html
	}
	if html, ok := cvStore["de"]; ok {
		return html
	}
	return ""
}
