package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/contact"
	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/internal/o11y"
	"github.com/srkn0/main/internal/templates/templatetargets"
	"github.com/srkn0/main/internal/templates/ui/components"
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
	logger     *slog.Logger
}

func (h *handlers) landing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	vm := pages.HomeViewModel{
		LatestPosts:      h.posts.Latest(locale, 5),
		FeaturedProjects: h.projects.Featured(locale, 4),
		TopTags:          h.posts.TagCounts(locale),
	}
	h.writeLayout(w, r, i18npkg.T(ctx, "page_home"), pages.LandingPage(ctx, vm))
}

func (h *handlers) blogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	archive := h.posts.ArchiveByYear(locale)
	h.writeLayout(w, r, i18npkg.T(ctx, "posts_title"), pages.BlogList(ctx, archive))
}

func (h *handlers) blogPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	post, ok := h.posts.Get(slug, i18npkg.GetLocale(ctx))
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.writeLayout(w, r, post.Title, pages.BlogPost(ctx, post))
}

func (h *handlers) projectList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	groups := h.projects.Grouped(locale)
	h.writeLayout(w, r, i18npkg.T(ctx, "projects_title"), pages.ProjectList(ctx, groups))
}

func (h *handlers) projectDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	project, ok := h.projects.Get(slug, i18npkg.GetLocale(ctx))
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.writeLayout(w, r, project.Title, pages.ProjectDetail(ctx, project))
}

func (h *handlers) cvPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := i18npkg.GetLocale(ctx)
	h.writeLayout(w, r, i18npkg.T(ctx, "nav_cv"), pages.CV(ctx, h.cv.Get(locale)))
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
	h.writeLayout(w, r, i18npkg.T(ctx, "nav_contact"), pages.Contact(ctx))
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
			o11y.LogEvent(ctx, h.logger, o11y.EventContactValidation,
				slog.String("email", form.Email),
			)
			writeStatus(w, http.StatusBadRequest, i18npkg.T(ctx, "contact_error_validation"), true)
			return
		}
		o11y.LogEvent(ctx, h.logger, o11y.EventContactSendFailed,
			slog.String("err", err.Error()),
		)
		writeStatus(w, http.StatusInternalServerError, i18npkg.T(ctx, "contact_error_send"), true)
		return
	}
	o11y.LogEvent(ctx, h.logger, o11y.EventContactSubmit,
		slog.String("name", form.Name),
		slog.String("email", form.Email),
		slog.String("subject", form.Subject),
	)
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
	ctx := r.Context()
	from := i18npkg.GetLocale(ctx)
	lang := r.URL.Query().Get("set")
	if !i18npkg.IsValidLang(lang) {
		lang = "de"
	}
	i18npkg.SetLangCookie(w, lang)
	o11y.LogEvent(ctx, h.logger, o11y.EventLangSwitch,
		slog.String("from", from),
		slog.String("to", lang),
	)
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

func (h *handlers) writeLayout(w http.ResponseWriter, r *http.Request, title string, component templ.Component) {
	ctx := i18npkg.WithPageTitle(r.Context(), title)
	locale := i18npkg.GetLocale(ctx)
	ctx = components.WithLayoutData(ctx, components.LayoutViewModel{
		CurrentPath:   r.URL.Path,
		LatestPosts: h.posts.Latest(locale, 2),
		ProjectLinks:  h.projects.Latest(locale, 3),
		ExternalLinks: h.externalLinks(locale),
	})
	if err := render.Layout(ctx, r, w, templatetargets.Main, component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handlers) externalLinks(locale string) []components.ExternalLink {
	projects := h.projects.GetAll(locale)
	for _, project := range projects {
		if project.Repo == "" {
			continue
		}
		parsed, err := url.Parse(project.Repo)
		if err != nil {
			continue
		}
		if strings.EqualFold(parsed.Host, "github.com") {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) > 0 && parts[0] != "" {
				return []components.ExternalLink{{
					Label: "GitHub",
					URL:   "https://github.com/" + parts[0],
				}}
			}
		}
	}
	return nil
}
