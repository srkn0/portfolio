package content

import (
	"bytes"
	"fmt"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

const defaultLocale = "de"

func isSupportedLocale(locale string) bool {
	return locale == "de" || locale == "en"
}

func convertMarkdown(src []byte) (htmlContent string, meta map[string]any, err error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			&frontmatter.Extender{Mode: frontmatter.SetMetadata},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	root := md.Parser().Parse(text.NewReader(src))
	doc := root.OwnerDocument()
	meta = doc.Meta()
	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, root); err != nil {
		return "", nil, err
	}
	return buf.String(), meta, nil
}

func parseMeta(meta map[string]any) (title, description string, tags []string, date time.Time, err error) {
	title, _ = meta["title"].(string)
	description, _ = meta["description"].(string)
	if rawTags, ok := meta["tags"].([]any); ok {
		for _, t := range rawTags {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
	}
	if rawDate, ok := meta["date"].(time.Time); ok {
		date = rawDate
	} else if rawDateStr, ok := meta["date"].(string); ok {
		date, err = time.Parse("2006-01-02", rawDateStr)
		if err != nil {
			return "", "", nil, time.Time{}, fmt.Errorf("invalid date format %q: %w", rawDateStr, err)
		}
	}
	return title, description, tags, date, nil
}
