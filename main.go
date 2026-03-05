package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/templates/ui/layouts"
	"github.com/srkn0/main/internal/templates/ui/pages"
)

//go:embed all:public
var publicFS embed.FS

func main() {
	publicSub, _ := fs.Sub(publicFS, "public")
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		layouts.Index().Render(ctx, w)
		pages.LandingPage().Render(ctx, w)
	})
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.FS(publicSub))))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
