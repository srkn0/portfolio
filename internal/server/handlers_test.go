package server

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/pkg/i18n"
)

// newTestHandlers wires up real stores from in-memory markdown so the templ render path actually runs.
func newTestHandlers(t *testing.T) *handlers {
	t.Helper()
	if err := i18n.Init(i18nFS()); err != nil {
		t.Fatalf("i18n init: %v", err)
	}
	posts, err := content.LoadPosts(filesFS(map[string]string{
		"hello/de.md": "---\ntitle: \"Hallo\"\ndate: 2026-05-01\n---\n\nbody",
	}))
	if err != nil {
		t.Fatalf("LoadPosts: %v", err)
	}
	projects, err := content.LoadProjects(filesFS(map[string]string{
		"demo/de.md": "---\ntitle: \"Demo\"\ndate: 2026-05-01\n---\n",
	}))
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	cv, err := content.LoadCV(filesFS(map[string]string{
		"cv.de.md": "# CV\n",
	}))
	if err != nil {
		t.Fatalf("LoadCV: %v", err)
	}
	return &handlers{posts: posts, projects: projects, cv: cv}
}

func filesFS(files map[string]string) fs.FS {
	m := fstest.MapFS{}
	for path, body := range files {
		m[path] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

func i18nFS() fs.FS {
	de := `{
		"site_title": "SK",
		"nav_blog": "Blog",
		"blog_title": "Blog",
		"projects_title": "Projekte"
	}`
	return fstest.MapFS{
		"de.json": &fstest.MapFile{Data: []byte(de)},
	}
}

func TestParsePage(t *testing.T) {
	cases := map[string]int{
		"":    1,
		"abc": 1,
		"-1":  1,
		"0":   1,
		"1":   1,
		"42":  42,
	}
	for in, want := range cases {
		if got := parsePage(in); got != want {
			t.Errorf("parsePage(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestBlogPost_unknownSlugReturns404(t *testing.T) {
	h := newTestHandlers(t)
	rec := httptest.NewRecorder()
	req := withChiURLParam(httptest.NewRequest(http.MethodGet, "/blog/missing", nil), "slug", "missing")
	h.blogPost(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestProjectDetail_unknownSlugReturns404(t *testing.T) {
	h := newTestHandlers(t)
	rec := httptest.NewRecorder()
	req := withChiURLParam(httptest.NewRequest(http.MethodGet, "/projects/missing", nil), "slug", "missing")
	h.projectDetail(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestSetLanguage_setsCookie(t *testing.T) {
	h := newTestHandlers(t)
	rec := httptest.NewRecorder()
	h.setLanguage(rec, httptest.NewRequest(http.MethodGet, "/lang?set=en", nil))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "en" {
		t.Fatalf("expected single lang=en cookie, got %v", cookies)
	}
}

func TestSetLanguage_unknownLangFallsBackToDE(t *testing.T) {
	h := newTestHandlers(t)
	rec := httptest.NewRecorder()
	h.setLanguage(rec, httptest.NewRequest(http.MethodGet, "/lang?set=fr", nil))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "de" {
		t.Fatalf("expected fallback to de, got %v", cookies)
	}
}

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
