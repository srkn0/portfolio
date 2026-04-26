package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/internal/templates/templatetargets"
	"github.com/srkn0/main/internal/templates/ui/pages"
	i18npkg "github.com/srkn0/main/pkg/i18n"
	"github.com/srkn0/main/pkg/util/render"
)

//go:embed all:public
var publicFS embed.FS

//go:embed all:data
var dataFS embed.FS

//go:embed all:locales
var localesFS embed.FS

func main() {
	publicSub, _ := fs.Sub(publicFS, "public")
	postsFS, _ := fs.Sub(dataFS, "data/posts")
	projectsFS, _ := fs.Sub(dataFS, "data/projects")
	cvFS, _ := fs.Sub(dataFS, "data/cv")
	localesSub, _ := fs.Sub(localesFS, "locales")

	i18npkg.Init(localesSub)
	posts, err := content.LoadPosts(postsFS)
	if err != nil { log.Fatal(err) }
	projects, err := content.LoadProjects(projectsFS)
	if err != nil { log.Fatal(err) }
	cv, err := content.LoadCV(cvFS)
	if err != nil { log.Fatal(err) }

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(i18npkg.Middleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.LandingPage(r.Context()))
	})
	r.Get("/blog", func(w http.ResponseWriter, r *http.Request) {
		ps, _, _ := posts.GetAll(1, 10, i18npkg.GetLocale(r.Context()))
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.BlogList(r.Context(), ps))
	})
	r.Get("/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		post, ok := posts.Get(slug, i18npkg.GetLocale(r.Context()))
		if !ok { http.NotFound(w, r); return }
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.BlogPost(r.Context(), post))
	})
	r.Get("/projects", func(w http.ResponseWriter, r *http.Request) {
		ps := projects.GetAll(i18npkg.GetLocale(r.Context()))
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.ProjectList(r.Context(), ps))
	})
	r.Get("/projects/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		project, ok := projects.Get(slug, i18npkg.GetLocale(r.Context()))
		if !ok { http.NotFound(w, r); return }
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.ProjectDetail(r.Context(), project))
	})
	r.Get("/cv", func(w http.ResponseWriter, r *http.Request) {
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.CV(r.Context(), cv.Get(i18npkg.GetLocale(r.Context()))))
	})
	r.Get("/lang", func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("set")
		if !i18npkg.IsValidLang(lang) { lang = "de" }
		i18npkg.SetLangCookie(w, lang)
		w.Header().Set("HX-Redirect", r.Header.Get("HX-Current-URL"))
		w.WriteHeader(http.StatusOK)
	})
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.FS(publicSub))))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
