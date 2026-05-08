package server

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/content"
	i18npkg "github.com/srkn0/main/pkg/i18n"
)

func Run(dataFS, publicFS fs.FS) error {
	return RunWithConfig(dataFS, publicFS, DefaultConfig())
}

func RunWithConfig(dataFS, publicFS fs.FS, cfg Config) error {
	posts, projects, cv, err := loadStores(dataFS)
	if err != nil {
		return err
	}

	h := &handlers{posts: posts, projects: projects, cv: cv}
	router := newRouter(h, publicFS)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return runWithShutdown(srv, cfg.ShutdownTimeout)
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

func newRouter(h *handlers, publicFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(i18npkg.Middleware)

	r.Get("/", h.landing)
	r.Get("/blog", h.blogList)
	r.Get("/blog/{slug}", h.blogPost)
	r.Get("/projects", h.projectList)
	r.Get("/projects/{slug}", h.projectDetail)
	r.Get("/cv", h.cvPage)
	r.Get("/cv/print", h.cvPrint)
	r.Get("/contact", h.contact)
	r.Get("/lang", h.setLanguage)

	r.Handle("/public/*", publicFileServer(publicFS))

	return r
}

func publicFileServer(publicFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(publicFS))
	return http.StripPrefix("/public/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(w, r)
	}))
}

func runWithShutdown(srv *http.Server, shutdownTimeout time.Duration) error {
	idleConnsClosed := make(chan struct{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-idleConnsClosed
	return nil
}
