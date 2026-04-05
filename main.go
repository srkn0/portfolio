package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	markdowntohtml "github.com/srkn0/main/internal/markdown_to_html"
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
	localesSub, _ := fs.Sub(localesFS, "locales")

	i18npkg.Init(localesSub)
	if err := markdowntohtml.LoadPosts(postsFS); err != nil {
		log.Fatal(err)
	}
	if err := markdowntohtml.LoadProjects(projectsFS); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(i18npkg.Middleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.LandingPage(r.Context()))
	})
	r.Get("/blog", func(w http.ResponseWriter, r *http.Request) {
		posts := markdowntohtml.GetAllPosts(i18npkg.GetLocale(r.Context()))
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.BlogList(r.Context(), posts))
	})
	r.Get("/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		post, ok := markdowntohtml.GetPost(slug, i18npkg.GetLocale(r.Context()))
		if !ok { http.NotFound(w, r); return }
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.BlogPost(r.Context(), post))
	})
	r.Get("/projects", func(w http.ResponseWriter, r *http.Request) {
		projects := markdowntohtml.GetAllProjects(i18npkg.GetLocale(r.Context()))
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.ProjectList(r.Context(), projects))
	})
	r.Get("/projects/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		project, ok := markdowntohtml.GetProject(slug, i18npkg.GetLocale(r.Context()))
		if !ok { http.NotFound(w, r); return }
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.ProjectDetail(r.Context(), project))
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
