package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/srkn0/main/internal/templates/templatetargets"
	"github.com/srkn0/main/internal/templates/ui/pages"
	"github.com/srkn0/main/pkg/util/render"
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
		render.Layout(r.Context(), r, w, templatetargets.Main, pages.LandingPage(r.Context()))
	})
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.FS(publicSub))))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
