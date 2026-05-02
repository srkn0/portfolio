package server

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/internal/templates/templatetargets"
	"github.com/srkn0/main/internal/templates/ui/pages"
	i18npkg "github.com/srkn0/main/pkg/i18n"
	"github.com/srkn0/main/pkg/util/render"
)

const postsPerPage = 10

type handlers struct {
	posts    *content.PostStore
	projects *content.ProjectStore
	cv       *content.CVStore
}

func (h *handlers) landing(w http.ResponseWriter, r *http.Request) {
	locale := i18npkg.GetLocale(r.Context())
	latestPosts, _, _ := h.posts.GetAll(1, postsPerPage, locale)
	latestProjects := h.projects.GetAll(locale)
	writeLayout(w, r, pages.LandingPage(r.Context(), latestPosts, latestProjects))
}

func (h *handlers) blogList(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r.URL.Query().Get("page"))
	locale := i18npkg.GetLocale(r.Context())
	posts, totalPages, _ := h.posts.GetAll(page, postsPerPage, locale)
	writeLayout(w, r, pages.BlogList(r.Context(), posts, page, totalPages))
}

func (h *handlers) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	post, ok := h.posts.Get(slug, i18npkg.GetLocale(r.Context()))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeLayout(w, r, pages.BlogPost(r.Context(), post))
}

func (h *handlers) projectList(w http.ResponseWriter, r *http.Request) {
	locale := i18npkg.GetLocale(r.Context())
	projects := h.projects.GetAll(locale)
	writeLayout(w, r, pages.ProjectList(r.Context(), projects))
}

func (h *handlers) projectDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	project, ok := h.projects.Get(slug, i18npkg.GetLocale(r.Context()))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeLayout(w, r, pages.ProjectDetail(r.Context(), project))
}

func (h *handlers) cvPage(w http.ResponseWriter, r *http.Request) {
	locale := i18npkg.GetLocale(r.Context())
	writeLayout(w, r, pages.CV(r.Context(), h.cv.Get(locale)))
}

func (h *handlers) cvPrint(w http.ResponseWriter, r *http.Request) {
	locale := i18npkg.GetLocale(r.Context())
	html := h.cv.Get(locale)
	if err := render.Layout(r.Context(), r, w, templatetargets.Print, pages.CVPrint(r.Context(), html)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handlers) contact(w http.ResponseWriter, r *http.Request) {
	writeLayout(w, r, pages.Contact(r.Context()))
}

func (h *handlers) setLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("set")
	if !i18npkg.IsValidLang(lang) {
		lang = "de"
	}
	i18npkg.SetLangCookie(w, lang)
	w.Header().Set("HX-Redirect", r.Header.Get("HX-Current-URL"))
	w.WriteHeader(http.StatusOK)
}

func parsePage(raw string) int {
	if raw == "" {
		return 1
	}
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func writeLayout(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := render.Layout(r.Context(), r, w, templatetargets.Main, component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
