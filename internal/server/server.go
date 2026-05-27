package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/srkn0/main/internal/contact"
	"github.com/srkn0/main/internal/content"
	"github.com/srkn0/main/internal/o11y"
	i18npkg "github.com/srkn0/main/pkg/i18n"
)

// Deps bundles cross-cutting dependencies the server needs at runtime.
// Bundling them keeps Run's signature stable as observability grows.
type Deps struct {
	Logger     *slog.Logger
	Recorder   o11y.Recorder
	ContactSvc *contact.Service
	Version    string // build version, stamped on the startup log + build_info metric
	Commit     string // build commit, same purpose
}

// Run boots the HTTP server with logging, metrics, traces and lifecycle
// signals wired in. The server stops on SIGINT/SIGTERM and drains in-flight
// requests up to cfg.ShutdownTimeout.
func Run(dataFS, publicFS fs.FS, cfg Config, deps Deps) error {
	posts, projects, cv, err := loadStores(dataFS)
	if err != nil {
		return err
	}

	_, _, postCount := posts.GetAll(1, 1, i18npkg.DefaultLocale)
	projectCount := len(projects.GetAll(i18npkg.DefaultLocale))
	deps.Recorder.SetPostCount(postCount)
	deps.Recorder.SetProjectCount(projectCount)

	h := &handlers{
		posts:      posts,
		projects:   projects,
		cv:         cv,
		contactSvc: deps.ContactSvc,
		logger:     deps.Logger,
	}

	// Readiness: stores are loaded synchronously before this point, so the
	// app is ready immediately. Extend this when downstream deps are added.
	ready := func() error { return nil }

	router := newRouter(h, publicFS, deps, ready)
	traced := o11y.TraceMiddleware(router)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      traced,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	printBanner(addr, postCount, projectCount)
	o11y.LogStartup(deps.Logger, o11y.StartupInfo{
		Version:      deps.Version,
		Commit:       deps.Commit,
		Addr:         addr,
		PostCount:    postCount,
		ProjectCount: projectCount,
	})

	return runWithShutdown(srv, cfg.ShutdownTimeout, deps.Logger)
}

func loadStores(dataFS fs.FS) (*content.PostStore, *content.ProjectStore, *content.CVStore, error) {
	postsFS, err := fs.Sub(dataFS, "posts")
	if err != nil {
		return nil, nil, nil, err
	}
	posts, err := content.LoadPosts(postsFS)
	if err != nil {
		return nil, nil, nil, err
	}

	projectsFS, err := fs.Sub(dataFS, "projects")
	if err != nil {
		return nil, nil, nil, err
	}
	projects, err := content.LoadProjects(projectsFS)
	if err != nil {
		return nil, nil, nil, err
	}

	cvFS, err := fs.Sub(dataFS, "cv")
	if err != nil {
		return nil, nil, nil, err
	}
	cv, err := content.LoadCV(cvFS)
	if err != nil {
		return nil, nil, nil, err
	}

	return posts, projects, cv, nil
}

func newRouter(h *handlers, publicFS fs.FS, deps Deps, ready o11y.ReadinessFunc) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Infra endpoints — no i18n, no request log (Prom scrapes /metrics every
	// few seconds; logging that would drown the user-traffic signal).
	r.Handle("/metrics", o11y.MetricsHandler())
	r.Handle("/healthz", o11y.LivenessHandler())
	r.Handle("/readyz", o11y.ReadinessHandler(ready))

	// App routes — structured request logging, metrics, locale resolution.
	r.Group(func(r chi.Router) {
		r.Use(o11y.RequestLog(deps.Logger))
		r.Use(o11y.MetricsMiddleware(deps.Recorder))
		r.Use(i18npkg.Middleware)

		r.Get("/", h.landing)
		r.Get("/blog", h.blogList)
		r.Get("/blog/{slug}", h.blogPost)
		r.Get("/projects", h.projectList)
		r.Get("/projects/{slug}", h.projectDetail)
		r.Get("/cv", h.cvPage)
		r.Get("/cv/print", h.cvPrint)
		r.Get("/contact", h.contact)
		r.Post("/contact", h.contactSubmit)
		r.Get("/lang", h.setLanguage)

		r.Handle("/public/*", publicFileServer(publicFS))
	})

	return r
}

func publicFileServer(publicFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(publicFS))
	return http.StripPrefix("/public/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(w, r)
	}))
}

func runWithShutdown(srv *http.Server, shutdownTimeout time.Duration, logger *slog.Logger) error {
	idleConnsClosed := make(chan struct{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		drainStart := time.Now()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		o11y.LogShutdown(logger, time.Since(drainStart))
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-idleConnsClosed
	return nil
}
