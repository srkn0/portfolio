package markdowntohtml

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type PostSummary struct {
	Title       string
	Description string
	Tags        []string
	Slug        string
	Date        time.Time
}

type Post struct {
	PostSummary
	HTMLContent string
}

var store map[string]map[string]Post

func LoadPosts(fsys fs.FS) error {
	store = make(map[string]map[string]Post)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading posts directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		src, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return fmt.Errorf("reading %s: %w", entry.Name(), err)
		}
		htmlContent, meta, err := convertMarkdown(src)
		if err != nil {
			return fmt.Errorf("converting %s: %w", entry.Name(), err)
		}
		title, description, tags, date, err := parseMeta(meta)
		if err != nil {
			return fmt.Errorf("parsing frontmatter for %s: %w", entry.Name(), err)
		}
		slug, locale := parseFilename(entry.Name())
		if store[slug] == nil {
			store[slug] = make(map[string]Post)
		}
		store[slug][locale] = Post{
			PostSummary: PostSummary{Title: title, Description: description, Tags: tags, Slug: slug, Date: date},
			HTMLContent: htmlContent,
		}
	}
	return nil
}

func GetAllPosts(locale string) []PostSummary {
	summaries := make([]PostSummary, 0, len(store))
	for _, locales := range store {
		p := pickLocale(locales, locale)
		if p == nil {
			continue
		}
		summaries = append(summaries, p.PostSummary)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Date.After(summaries[j].Date) })
	return summaries
}

func GetPost(slug, locale string) (Post, bool) {
	locales, ok := store[slug]
	if !ok {
		return Post{}, false
	}
	p := pickLocale(locales, locale)
	if p == nil {
		return Post{}, false
	}
	return *p, true
}

func pickLocale(locales map[string]Post, locale string) *Post {
	if p, ok := locales[locale]; ok { return &p }
	if p, ok := locales["de"]; ok { return &p }
	for _, p := range locales { return &p }
	return nil
}
