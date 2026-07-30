package content

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type ProjectSummary struct {
	Title       string
	Description string
	Tags        []string
	Slug        string
	Date        time.Time
	Image       string
	Repo        string
	Demo        string
}

type Project struct {
	ProjectSummary
	HTMLContent string
}

type ProjectStore struct {
	byLocale map[string]map[string]Project
}

func LoadProjects(fsys fs.FS) (*ProjectStore, error) {
	store := &ProjectStore{byLocale: make(map[string]map[string]Project)}

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("reading projects directory: %w", err)
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

			image, _ := meta["image"].(string)
			repo, _ := meta["repo"].(string)
			demo, _ := meta["demo"].(string)

			if store.byLocale[slug] == nil {
				store.byLocale[slug] = make(map[string]Project)
			}
			store.byLocale[slug][locale] = Project{
				ProjectSummary: ProjectSummary{
					Title:       title,
					Description: description,
					Tags:        tags,
					Slug:        slug,
					Date:        date,
					Image:       image,
					Repo:        repo,
					Demo:        demo,
				},
				HTMLContent: htmlContent,
			}
		}
	}

	return store, nil
}

func (s *ProjectStore) GetAll(locale string) []ProjectSummary {
	summaries := make([]ProjectSummary, 0, len(s.byLocale))
	for _, locales := range s.byLocale {
		p := pickProjectLocale(locales, locale)
		if p == nil {
			continue
		}
		summaries = append(summaries, p.ProjectSummary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Date.After(summaries[j].Date)
	})

	return summaries
}

func (s *ProjectStore) Latest(locale string, limit int) []ProjectSummary {
	projects := s.GetAll(locale)
	if limit <= 0 || limit >= len(projects) {
		return projects
	}
	return projects[:limit]
}

func (s *ProjectStore) Get(slug, locale string) (Project, bool) {
	locales, ok := s.byLocale[slug]
	if !ok {
		return Project{}, false
	}
	p := pickProjectLocale(locales, locale)
	if p == nil {
		return Project{}, false
	}
	return *p, true
}

func pickProjectLocale(locales map[string]Project, locale string) *Project {
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
