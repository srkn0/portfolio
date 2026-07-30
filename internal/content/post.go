package content

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

type PostArchiveYear struct {
	Year  int
	Posts []PostSummary
}

type TagCount struct {
	Name  string
	Count int
}

type Post struct {
	PostSummary
	HTMLContent string
}

type PostStore struct {
	byLocale map[string]map[string]Post
}

func LoadPosts(fsys fs.FS) (*PostStore, error) {
	store := &PostStore{byLocale: make(map[string]map[string]Post)}

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("reading posts directory: %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		slug := dir.Name()

		files, err := fs.ReadDir(fsys, slug)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", slug, err)
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}
			locale := strings.TrimSuffix(file.Name(), ".md")
			if !isSupportedLocale(locale) {
				continue
			}

			path := slug + "/" + file.Name()
			src, err := fs.ReadFile(fsys, path)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", path, err)
			}

			htmlContent, meta, err := convertMarkdown(src)
			if err != nil {
				return nil, fmt.Errorf("converting %s: %w", path, err)
			}

			title, description, tags, date, err := parseMeta(meta)
			if err != nil {
				return nil, fmt.Errorf("parsing frontmatter for %s: %w", path, err)
			}

			if store.byLocale[slug] == nil {
				store.byLocale[slug] = make(map[string]Post)
			}
			store.byLocale[slug][locale] = Post{
				PostSummary: PostSummary{
					Title:       title,
					Description: description,
					Tags:        tags,
					Slug:        slug,
					Date:        date,
				},
				HTMLContent: htmlContent,
			}
		}
	}

	return store, nil
}

func (s *PostStore) GetAll(page, perPage int, locale string) ([]PostSummary, int, int) {
	summaries := s.List(locale)

	totalPosts := len(summaries)
	totalPages := 1
	if totalPosts > perPage {
		totalPages = (totalPosts + perPage - 1) / perPage
	}

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * perPage
	end := min(start+perPage, totalPosts)
	if start > totalPosts {
		return []PostSummary{}, totalPages, totalPosts
	}

	return summaries[start:end], totalPages, totalPosts
}

func (s *PostStore) List(locale string) []PostSummary {
	summaries := make([]PostSummary, 0, len(s.byLocale))
	for _, locales := range s.byLocale {
		p := pickLocale(locales, locale)
		if p == nil {
			continue
		}
		summaries = append(summaries, p.PostSummary)
	}

	sortPostSummaries(summaries)

	return summaries
}

func (s *PostStore) Latest(locale string, limit int) []PostSummary {
	posts := s.List(locale)
	if limit <= 0 || limit >= len(posts) {
		return posts
	}
	return posts[:limit]
}

func (s *PostStore) ArchiveByYear(locale string) []PostArchiveYear {
	posts := s.List(locale)
	years := make([]PostArchiveYear, 0)
	yearIndex := make(map[int]int)

	for _, post := range posts {
		year := post.Date.Year()
		idx, ok := yearIndex[year]
		if !ok {
			idx = len(years)
			yearIndex[year] = idx
			years = append(years, PostArchiveYear{Year: year})
		}
		years[idx].Posts = append(years[idx].Posts, post)
	}

	return years
}

func (s *PostStore) TagCounts(locale string) []TagCount {
	counts := make(map[string]int)
	for _, post := range s.List(locale) {
		for _, tag := range post.Tags {
			counts[tag]++
		}
	}

	tags := make([]TagCount, 0, len(counts))
	for name, count := range counts {
		tags = append(tags, TagCount{Name: name, Count: count})
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Count == tags[j].Count {
			return strings.ToLower(tags[i].Name) < strings.ToLower(tags[j].Name)
		}
		return tags[i].Count > tags[j].Count
	})

	return tags
}

func (s *PostStore) Get(slug, locale string) (Post, bool) {
	locales, ok := s.byLocale[slug]
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
	if p, ok := locales[locale]; ok {
		return &p
	}
	if p, ok := locales[defaultLocale]; ok {
		return &p
	}
	for _, p := range locales {
		return &p
	}
	return nil
}

func sortPostSummaries(posts []PostSummary) {
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Date.Equal(posts[j].Date) {
			return strings.ToLower(posts[i].Title) < strings.ToLower(posts[j].Title)
		}
		return posts[i].Date.After(posts[j].Date)
	})
}
