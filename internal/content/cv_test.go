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

// The AST transformer is what justifies cv having its own loader at all.
// It hangs Tailwind classes on the rendered HTML so the /cv/print stylesheet
// can target them. If someone refactors the transformer away, the CV print
// layout silently loses its styling.

func TestASTTransformer_headingClasses(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"de.md": "# H1\n\n## H2\n\n### H3\n",
	}))
	html := store.Get("de")

	if !strings.Contains(html, "text-3xl") {
		t.Errorf("h1 missing text-3xl: %s", html)
	}
	if !strings.Contains(html, "text-lg") {
		t.Errorf("h2 missing text-lg: %s", html)
	}
	if !strings.Contains(html, "text-base") {
		t.Errorf("h3 missing text-base: %s", html)
	}
}

func TestASTTransformer_listsAndBreaks(t *testing.T) {
	store, _ := LoadCV(cvFS(map[string]string{
		"de.md": "- one\n- two\n\n1. first\n2. second\n\n---\n",
	}))
	html := store.Get("de")

	if !strings.Contains(html, "list-disc") {
		t.Errorf("unordered list missing list-disc: %s", html)
	}
	if !strings.Contains(html, "list-decimal") {
		t.Errorf("ordered list missing list-decimal: %s", html)
	}
	if !strings.Contains(html, "border-border") {
		t.Errorf("thematic break missing class: %s", html)
	}
}
