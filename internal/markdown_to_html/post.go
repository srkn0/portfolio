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
	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading posts directory: %w", err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		slug := dir.Name()
		files, err := fs.ReadDir(fsys, slug)
		if err != nil {
			return fmt.Errorf("reading %s: %w", slug, err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}
			locale := strings.TrimSuffix(file.Name(), ".md")
			if locale != "de" && locale != "en" {
				continue
			}
			path := slug + "/" + file.Name()
			src, err := fs.ReadFile(fsys, path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			htmlContent, meta, err := convertMarkdown(src)
			if err != nil {
				return fmt.Errorf("converting %s: %w", path, err)
			}
			title, description, tags, date, err := parseMeta(meta)
			if err != nil {
				return fmt.Errorf("parsing frontmatter for %s: %w", path, err)
			}
			if store[slug] == nil {
				store[slug] = make(map[string]Post)
			}
			store[slug][locale] = Post{
				PostSummary: PostSummary{Title: title, Description: description, Tags: tags, Slug: slug, Date: date},
				HTMLContent: htmlContent,
			}
		}
	}
	return nil
}

func GetAllPosts(locale string) []PostSummary {
	summaries := make([]PostSummary, 0, len(store))
	for _, locales := range store {
		p := pickLocale(locales, locale)
		if p == nil { continue }
		summaries = append(summaries, p.PostSummary)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Date.After(summaries[j].Date) })
	return summaries
}

func GetPost(slug, locale string) (Post, bool) {
	locales, ok := store[slug]
	if !ok { return Post{}, false }
	p := pickLocale(locales, locale)
	if p == nil { return Post{}, false }
	return *p, true
}

func pickLocale(locales map[string]Post, locale string) *Post {
	if p, ok := locales[locale]; ok { return &p }
	if p, ok := locales["de"]; ok { return &p }
	for _, p := range locales { return &p }
	return nil
}
