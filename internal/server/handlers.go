package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/contact"
	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/internal/templates/templatetargets"
	"github.com/srkn0/main/internal/templates/ui/pages"
	i18npkg "github.com/srkn0/main/pkg/i18n"
	"github.com/srkn0/main/pkg/util/render"
)

const postsPerPage = 10

type handlers struct {
	posts      *content.PostStore
	projects   *content.ProjectStore
	cv         *content.CVStore
	contactSvc *contact.Service
}

func (h *handlers) landing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	latestPosts, _, _ := h.posts.GetAll(1, postsPerPage, locale)
	latestProjects := h.projects.GetAll(locale)
	writeLayout(w, r, i18npkg.T(ctx, "page_home"), pages.LandingPage(ctx, latestPosts, latestProjects))
}

func (h *handlers) blogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := parsePage(r.URL.Query().Get("page"))
	locale := i18npkg.GetLocale(ctx)
	posts, totalPages, _ := h.posts.GetAll(page, postsPerPage, locale)
	writeLayout(w, r, i18npkg.T(ctx, "nav_blog"), pages.BlogList(ctx, posts, page, totalPages))
}

func (h *handlers) blogPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	post, ok := h.posts.Get(slug, i18npkg.GetLocale(ctx))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeLayout(w, r, post.Title, pages.BlogPost(ctx, post))
}

func (h *handlers) projectList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	projects := h.projects.GetAll(locale)
	writeLayout(w, r, i18npkg.T(ctx, "nav_projects"), pages.ProjectList(ctx, projects))
}

func (h *handlers) projectDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	project, ok := h.projects.Get(slug, i18npkg.GetLocale(ctx))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeLayout(w, r, project.Title, pages.ProjectDetail(ctx, project))
}

func (h *handlers) cvPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	writeLayout(w, r, i18npkg.T(ctx, "nav_cv"), pages.CV(ctx, h.cv.Get(locale)))
}

func (h *handlers) cvPrint(w http.ResponseWriter, r *http.Request) {
	ctx := i18npkg.WithPageTitle(r.Context(), i18npkg.T(r.Context(), "nav_cv"))
	locale := i18npkg.GetLocale(ctx)
	html := h.cv.Get(locale)
	if err := render.Layout(ctx, r, w, templatetargets.Print, pages.CVPrint(ctx, html)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handlers) contact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writeLayout(w, r, i18npkg.T(ctx, "nav_contact"), pages.Contact(ctx))
}

func (h *handlers) contactSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		writeStatus(w, http.StatusBadRequest, i18npkg.T(ctx, "contact_error_validation"), true)
		return
	}
	form := contact.Form{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}
	if err := h.contactSvc.Send(ctx, form); err != nil {
		if errors.Is(err, contact.ErrInvalidForm) {
			writeStatus(w, http.StatusBadRequest, i18npkg.T(ctx, "contact_error_validation"), true)
			return
		}
		log.Printf("contact: send failed: %v", err)
		writeStatus(w, http.StatusInternalServerError, i18npkg.T(ctx, "contact_error_send"), true)
		return
	}
	writeStatus(w, http.StatusOK, i18npkg.T(ctx, "contact_success"), false)
}

func writeStatus(w http.ResponseWriter, status int, msg string, isError bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	class := "text-foreground"
	if isError {
		class = "text-destructive"
	}
	fmt.Fprintf(w, `<p class="%s">%s</p>`, class, templ.EscapeString(msg))
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

func writeLayout(w http.ResponseWriter, r *http.Request, title string, component templ.Component) {
	ctx := i18npkg.WithPageTitle(r.Context(), title)
	if err := render.Layout(ctx, r, w, templatetargets.Main, component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
