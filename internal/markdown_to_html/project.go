package markdowntohtml

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

var projectStore map[string]map[string]Project

func LoadProjects(fsys fs.FS) error {
	projectStore = make(map[string]map[string]Project)
	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading projects directory: %w", err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() { continue }
		slug := dir.Name()
		files, err := fs.ReadDir(fsys, slug)
		if err != nil {
			return fmt.Errorf("reading %s: %w", slug, err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") { continue }
			locale := strings.TrimSuffix(file.Name(), ".md")
			if locale != "de" && locale != "en" { continue }
			path := slug + "/" + file.Name()
			src, err := fs.ReadFile(fsys, path)
			if err != nil { return fmt.Errorf("reading %s: %w", path, err) }
			htmlContent, meta, err := convertMarkdown(src)
			if err != nil { return fmt.Errorf("converting %s: %w", path, err) }
			title, description, tags, date, err := parseMeta(meta)
			if err != nil { return fmt.Errorf("parsing frontmatter for %s: %w", path, err) }
			image, _ := meta["image"].(string)
			repo, _ := meta["repo"].(string)
			demo, _ := meta["demo"].(string)
			if projectStore[slug] == nil {
				projectStore[slug] = make(map[string]Project)
			}
			projectStore[slug][locale] = Project{
				ProjectSummary: ProjectSummary{
					Title: title, Description: description, Tags: tags, Slug: slug, Date: date,
					Image: image, Repo: repo, Demo: demo,
				},
				HTMLContent: htmlContent,
			}
		}
	}
	return nil
}

func GetAllProjects(locale string) []ProjectSummary {
	out := make([]ProjectSummary, 0, len(projectStore))
	for _, locales := range projectStore {
		p := pickProjectLocale(locales, locale)
		if p == nil { continue }
		out = append(out, p.ProjectSummary)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out
}

func GetProject(slug, locale string) (Project, bool) {
	locales, ok := projectStore[slug]
	if !ok { return Project{}, false }
	p := pickProjectLocale(locales, locale)
	if p == nil { return Project{}, false }
	return *p, true
}

func pickProjectLocale(locales map[string]Project, locale string) *Project {
	if p, ok := locales[locale]; ok { return &p }
	if p, ok := locales["de"]; ok { return &p }
	for _, p := range locales { return &p }
	return nil
}
