package content

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func cvFS(files map[string]string) fs.FS {
	m := fstest.MapFS{}
	for path, body := range files {
		m[path] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

func TestExtractCVLocale(t *testing.T) {
	cases := map[string]string{
		"cv.de.md":         "de",
		"cv.en.md":         "en",
		"cv.md":            "de", // no dot before locale → defaults to de
		"cv.fr.md":         "fr",
		"some.name.de.md":  "de",
	}
	for in, want := range cases {
		if got := extractCVLocale(in); got != want {
			t.Errorf("extractCVLocale(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadCV_singleLocale(t *testing.T) {
	store, err := LoadCV(cvFS(map[string]string{
		"cv.de.md": "# CV\n\nHello.\n",
	}))
	if err != nil {
		t.Fatalf("LoadCV: %v", err)
	}
	got := store.Get("de")
	if !strings.Contains(got, "CV") {
		t.Errorf("expected CV heading in output, got: %s", got)
	}
}

func TestLoadCV_skipsUnsupportedLocale(t *testing.T) {
	store, err := LoadCV(cvFS(map[string]string{
		"cv.de.md": "# DE\n",
		"cv.fr.md": "# FR\n",
	}))
	if err != nil {
		t.Fatalf("LoadCV: %v", err)
	}
	if got := store.Get("fr"); got == "" {
		// fr was skipped, falling back to de
	} else if !strings.Contains(got, "DE") {
		t.Errorf("fr should not be loaded, got: %s", got)
	}
}

func TestCVStore_localeFallback(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"cv.de.md": "# DE only\n",
	}))
	// Asking for en falls back to de
	if got := store.Get("en"); !strings.Contains(got, "DE only") {
		t.Errorf("fallback failed, got: %s", got)
	}
}

func TestCVStore_returnsEmptyWhenNothingLoaded(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{}))
	if got := store.Get("de"); got != "" {
		t.Errorf("expected empty, got: %s", got)
	}
}

func TestASTTransformer_headingClasses(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"cv.de.md": "# H1\n\n## H2\n\n### H3\n\n#### H4\n",
	}))
	html := store.Get("de")

	if !strings.Contains(html, `<h1`) || !strings.Contains(html, "text-3xl") {
		t.Errorf("h1 missing or wrong class: %s", html)
	}
	if !strings.Contains(html, `<h2`) || !strings.Contains(html, "text-lg") {
		t.Errorf("h2 missing or wrong class: %s", html)
	}
	if !strings.Contains(html, `<h3`) || !strings.Contains(html, "text-base") {
		t.Errorf("h3 missing or wrong class: %s", html)
	}
}

func TestASTTransformer_paragraphsAndLists(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"cv.de.md": "Plain paragraph.\n\n- bullet one\n- bullet two\n\n1. ordered one\n2. ordered two\n",
	}))
	html := store.Get("de")

	if !strings.Contains(html, `<p class=`) {
		t.Errorf("paragraph missing class: %s", html)
	}
	if !strings.Contains(html, `<ul class=`) || !strings.Contains(html, "list-disc") {
		t.Errorf("unordered list missing class: %s", html)
	}
	if !strings.Contains(html, `<ol class=`) || !strings.Contains(html, "list-decimal") {
		t.Errorf("ordered list missing class: %s", html)
	}
}

func TestASTTransformer_thematicBreak(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"cv.de.md": "before\n\n---\n\nafter\n",
	}))
	html := store.Get("de")
	if !strings.Contains(html, `<hr`) || !strings.Contains(html, "border-border") {
		t.Errorf("thematic break missing class: %s", html)
	}
}

func TestHeadingClass_unknownLevel(t *testing.T) {
	// Level 4+ falls into the default branch
	if got := headingClass(4); !strings.Contains(got, "font-bold") {
		t.Errorf("level 4 class = %q, want font-bold", got)
	}
	if got := headingClass(6); !strings.Contains(got, "font-bold") {
		t.Errorf("level 6 class = %q, want font-bold", got)
	}
}

func TestListClass(t *testing.T) {
	if !strings.Contains(listClass(true), "list-decimal") {
		t.Error("ordered list missing list-decimal")
	}
	if !strings.Contains(listClass(false), "list-disc") {
		t.Error("unordered list missing list-disc")
	}
}
