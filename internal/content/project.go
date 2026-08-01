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
	Category    string
	Featured    int
}

type Project struct {
	ProjectSummary
	HTMLContent string
}

type ProjectGroup struct {
	Category string
	Projects []ProjectSummary
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
			category := metaString(meta, "category", "lab")
			featured := metaInt(meta, "featured")

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
					Category:    category,
					Featured:    featured,
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

	sortProjectSummaries(summaries)

	return summaries
}

func (s *ProjectStore) Featured(locale string, limit int) []ProjectSummary {
	projects := s.GetAll(locale)
	if limit <= 0 || limit >= len(projects) {
		return projects
	}
	return projects[:limit]
}

func (s *ProjectStore) Latest(locale string, limit int) []ProjectSummary {
	return s.Featured(locale, limit)
}

func (s *ProjectStore) Grouped(locale string) []ProjectGroup {
	projects := s.GetAll(locale)
	groupsByCategory := make(map[string][]ProjectSummary)
	seen := make(map[string]bool)
	for _, project := range projects {
		category := project.Category
		if category == "" {
			category = "lab"
		}
		groupsByCategory[category] = append(groupsByCategory[category], project)
		seen[category] = true
	}

	order := []string{"infrastructure", "platform", "template", "workstation", "lab", "wip"}
	groups := make([]ProjectGroup, 0, len(groupsByCategory))
	for _, category := range order {
		if !seen[category] {
			continue
		}
		groups = append(groups, ProjectGroup{Category: category, Projects: groupsByCategory[category]})
		delete(seen, category)
	}

	var remaining []string
	for category := range seen {
		remaining = append(remaining, category)
	}
	sort.Strings(remaining)
	for _, category := range remaining {
		groups = append(groups, ProjectGroup{Category: category, Projects: groupsByCategory[category]})
	}

	return groups
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

func sortProjectSummaries(projects []ProjectSummary) {
	sort.Slice(projects, func(i, j int) bool {
		left := projects[i]
		right := projects[j]
		if left.Featured > 0 || right.Featured > 0 {
			if left.Featured == 0 {
				return false
			}
			if right.Featured == 0 {
				return true
			}
			if left.Featured != right.Featured {
				return left.Featured < right.Featured
			}
		}
		return left.Date.After(right.Date)
	})
}

func metaString(meta map[string]any, key string, fallback string) string {
	if value, ok := meta[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func metaInt(meta map[string]any, key string) int {
	switch value := meta[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}
