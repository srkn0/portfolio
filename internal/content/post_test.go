package content_test

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/srkn0/main/internal/content"
)

func postFS(files map[string]string) fs.FS {
	m := fstest.MapFS{}
	for path, body := range files {
		m[path] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

func TestLoadPosts_singlePost(t *testing.T) {
	store, err := content.LoadPosts(postFS(map[string]string{
		"hello/de.md": "---\ntitle: \"Hello\"\ndate: 2026-05-01\n---\n\nbody",
	}))
	if err != nil {
		t.Fatalf("LoadPosts: %v", err)
	}
	post, ok := store.Get("hello", "de")
	if !ok {
		t.Fatal("post not found")
	}
	if post.Title != "Hello" {
		t.Errorf("title = %q", post.Title)
	}
	if post.Slug != "hello" {
		t.Errorf("slug = %q", post.Slug)
	}
	if !strings.Contains(post.HTMLContent, "body") {
		t.Errorf("html missing body: %s", post.HTMLContent)
	}
}

func TestLoadPosts_multipleLocales(t *testing.T) {
	store, err := content.LoadPosts(postFS(map[string]string{
		"hi/de.md": "---\ntitle: \"Hallo\"\ndate: 2026-05-01\n---\n",
		"hi/en.md": "---\ntitle: \"Hi\"\ndate: 2026-05-01\n---\n",
	}))
	if err != nil {
		t.Fatalf("LoadPosts: %v", err)
	}
	de, _ := store.Get("hi", "de")
	en, _ := store.Get("hi", "en")
	if de.Title != "Hallo" || en.Title != "Hi" {
		t.Errorf("de=%q en=%q", de.Title, en.Title)
	}
}

func TestLoadPosts_skipsUnsupportedLocale(t *testing.T) {
	// The loader must not store posts under "fr", even though the .md is on disk.
	// We assert this by requesting fr and confirming we fall back to the de version
	// instead of getting Bonjour back.
	store, err := content.LoadPosts(postFS(map[string]string{
		"hi/de.md": "---\ntitle: \"Hallo\"\ndate: 2026-05-01\n---\n",
		"hi/fr.md": "---\ntitle: \"Bonjour\"\ndate: 2026-05-01\n---\n",
	}))
	if err != nil {
		t.Fatalf("LoadPosts: %v", err)
	}
	got, ok := store.Get("hi", "fr")
	if !ok {
		t.Fatal("expected fallback to de")
	}
	if got.Title == "Bonjour" {
		t.Error("fr locale was loaded but should have been skipped")
	}
}

func TestLoadPosts_skipsLooseFiles(t *testing.T) {
	store, err := content.LoadPosts(postFS(map[string]string{
		"README.md": "loose file at root",
		"hi/de.md":  "---\ntitle: \"Hallo\"\ndate: 2026-05-01\n---\n",
		"hi/notes.txt": "ignored",
	}))
	if err != nil {
		t.Fatalf("LoadPosts: %v", err)
	}
	if _, ok := store.Get("hi", "de"); !ok {
		t.Error("expected hi post to load")
	}
}

func TestLoadPosts_invalidFrontmatterReturnsError(t *testing.T) {
	_, err := content.LoadPosts(postFS(map[string]string{
		"bad/de.md": "---\ndate: not-a-date\n---\n",
	}))
	if err == nil {
		t.Fatal("expected error for invalid frontmatter")
	}
}

func TestGetAll_sortedByDateDesc(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{
		"old/de.md": "---\ntitle: \"Old\"\ndate: 2026-01-01\n---\n",
		"new/de.md": "---\ntitle: \"New\"\ndate: 2026-05-01\n---\n",
		"mid/de.md": "---\ntitle: \"Mid\"\ndate: 2026-03-01\n---\n",
	}))
	posts, _, total := store.GetAll(1, 10, "de")
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if posts[0].Title != "New" || posts[1].Title != "Mid" || posts[2].Title != "Old" {
		t.Errorf("order = %v, want [New Mid Old]", []string{posts[0].Title, posts[1].Title, posts[2].Title})
	}
}

func TestGetAll_paginates(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{
		"a/de.md": "---\ntitle: \"A\"\ndate: 2026-05-01\n---\n",
		"b/de.md": "---\ntitle: \"B\"\ndate: 2026-05-02\n---\n",
		"c/de.md": "---\ntitle: \"C\"\ndate: 2026-05-03\n---\n",
		"d/de.md": "---\ntitle: \"D\"\ndate: 2026-05-04\n---\n",
		"e/de.md": "---\ntitle: \"E\"\ndate: 2026-05-05\n---\n",
	}))

	posts, totalPages, totalPosts := store.GetAll(1, 2, "de")
	if totalPosts != 5 {
		t.Errorf("totalPosts = %d, want 5", totalPosts)
	}
	if totalPages != 3 {
		t.Errorf("totalPages = %d, want 3", totalPages)
	}
	if len(posts) != 2 {
		t.Errorf("page 1 size = %d, want 2", len(posts))
	}

	posts, _, _ = store.GetAll(3, 2, "de")
	if len(posts) != 1 {
		t.Errorf("page 3 size = %d, want 1 (last partial page)", len(posts))
	}
}

func TestGetAll_clampsOutOfRangePage(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{
		"a/de.md": "---\ntitle: \"A\"\ndate: 2026-05-01\n---\n",
		"b/de.md": "---\ntitle: \"B\"\ndate: 2026-05-02\n---\n",
	}))

	// Page 0 clamps to 1
	posts, _, _ := store.GetAll(0, 10, "de")
	if len(posts) != 2 {
		t.Errorf("page 0 size = %d, want 2 (clamped)", len(posts))
	}

	// Page beyond max clamps to last
	posts, _, _ = store.GetAll(99, 10, "de")
	if len(posts) != 2 {
		t.Errorf("page 99 size = %d, want 2 (clamped)", len(posts))
	}
}

func TestGet_notFound(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{}))
	if _, ok := store.Get("missing", "de"); ok {
		t.Error("expected not found")
	}
}

func TestGet_localeFallback(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{
		"only-de/de.md": "---\ntitle: \"Nur DE\"\ndate: 2026-05-01\n---\n",
	}))
	// Asking for en falls back to de
	post, ok := store.Get("only-de", "en")
	if !ok {
		t.Fatal("expected fallback")
	}
	if post.Title != "Nur DE" {
		t.Errorf("title = %q", post.Title)
	}
}

func TestGetAll_emptyStore(t *testing.T) {
	store, _ := content.LoadPosts(postFS(map[string]string{}))
	posts, totalPages, totalPosts := store.GetAll(1, 10, "de")
	if len(posts) != 0 || totalPosts != 0 {
		t.Errorf("expected empty result, got %v %d", posts, totalPosts)
	}
	if totalPages != 1 {
		t.Errorf("totalPages = %d, want 1 (minimum)", totalPages)
	}
}
