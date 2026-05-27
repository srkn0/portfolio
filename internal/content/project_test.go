package content_test

import (
	"testing"

	"github.com/srkn0/main/internal/content"
)

func TestLoadProjects_extraFrontmatterFields(t *testing.T) {
	// Projects carry image/repo/demo on top of the post fields. Make sure
	// those round-trip through the loader.
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

func TestGetAllProjects_sortedByDateDesc(t *testing.T) {
	store, _ := content.LoadProjects(postFS(map[string]string{
		"old/de.md": "---\ntitle: \"Old\"\ndate: 2026-01-01\n---\n",
		"new/de.md": "---\ntitle: \"New\"\ndate: 2026-05-01\n---\n",
	}))
	all := store.GetAll("de")
	if len(all) != 2 || all[0].Title != "New" || all[1].Title != "Old" {
		t.Errorf("unexpected ordering: %v", all)
	}
}

func TestGetProject_localeFallback(t *testing.T) {
	store, _ := content.LoadProjects(postFS(map[string]string{
		"only-de/de.md": "---\ntitle: \"Nur DE\"\ndate: 2026-05-01\n---\n",
	}))
	p, ok := store.Get("only-de", "en")
	if !ok || p.Title != "Nur DE" {
		t.Errorf("fallback failed: ok=%v title=%q", ok, p.Title)
	}
}
