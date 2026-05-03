package server

import (
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/content"
	i18npkg "github.com/srkn0/main/pkg/i18n"
)

func Run(dataFS, publicFS fs.FS) error {
	posts, projects, cv, err := loadStores(dataFS)
	if err != nil { return err }

	h := &handlers{posts: posts, projects: projects, cv: cv}

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
	r.Get("/contact", h.contact)
	r.Get("/lang", h.setLanguage)

	fileServer := http.FileServer(http.FS(publicFS))
	r.Handle("/public/*", http.StripPrefix("/public/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(w, r)
	})))

	s := &http.Server{
		Addr: ":8080", Handler: r,
		ReadTimeout: 30 * time.Second, WriteTimeout: 90 * time.Second, IdleTimeout: 120 * time.Second,
	}
	log.Println("listening on :8080")
	return s.ListenAndServe()
}

func loadStores(dataFS fs.FS) (*content.PostStore, *content.ProjectStore, *content.CVStore, error) {
	postsFS, err := fs.Sub(dataFS, "posts")
	if err != nil { return nil, nil, nil, err }
	posts, err := content.LoadPosts(postsFS)
	if err != nil { return nil, nil, nil, err }

	projectsFS, err := fs.Sub(dataFS, "projects")
	if err != nil { return nil, nil, nil, err }
	projects, err := content.LoadProjects(projectsFS)
	if err != nil { return nil, nil, nil, err }

	cvFS, err := fs.Sub(dataFS, "cv")
	if err != nil { return nil, nil, nil, err }
	cv, err := content.LoadCV(cvFS)
	if err != nil { return nil, nil, nil, err }

	return posts, projects, cv, nil
}
