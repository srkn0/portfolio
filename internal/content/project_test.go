package content_test

import (
	"testing"

	"github.com/srkn0/main/internal/content"
)

// Project's loader is structurally identical to Post's, only the frontmatter
// has image/repo/demo on top. Test only that those round-trip; the rest is
// covered by post_test.go.
func TestLoadProjects_extraFrontmatterFields(t *testing.T) {
	store, err := content.LoadProjects(postFS(map[string]string{
		"foo/de.md": "---\ntitle: \"Foo\"\ndate: 2026-05-01\nrepo: https://example.com/repo\ndemo: https://example.com/demo\nimage: /img.png\n---\n\nbody",
	}))
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	p, ok := store.Get("foo", "de")
	if !ok {
		t.Fatal("project not found")
	}
	if p.Repo != "https://example.com/repo" {
		t.Errorf("repo = %q", p.Repo)
	}
	if p.Demo != "https://example.com/demo" {
		t.Errorf("demo = %q", p.Demo)
	}
	if p.Image != "/img.png" {
		t.Errorf("image = %q", p.Image)
	}
}

func TestProjectStoreLatest(t *testing.T) {
	store, err := content.LoadProjects(postFS(map[string]string{
		"old/de.md": "---\ntitle: \"Old\"\ndate: 2026-01-01\n---\n",
		"new/de.md": "---\ntitle: \"New\"\ndate: 2026-05-01\n---\n",
		"mid/de.md": "---\ntitle: \"Mid\"\ndate: 2026-03-01\n---\n",
	}))
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}

	projects := store.Latest("de", 2)
	if len(projects) != 2 {
		t.Fatalf("Latest size = %d, want 2", len(projects))
	}
	if projects[0].Title != "New" || projects[1].Title != "Mid" {
		t.Fatalf("Latest order = %v", []string{projects[0].Title, projects[1].Title})
	}
}
