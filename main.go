package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/srkn0/main/internal/server"
	"github.com/srkn0/main/pkg/i18n"
)

//go:embed all:locales
var localesFS embed.FS

//go:embed all:data
var dataFS embed.FS

//go:embed all:public
var publicFS embed.FS

func main() {
	localesSub := mustSub(localesFS, "locales")
	dataSub := mustSub(dataFS, "data")
	publicSub := mustSub(publicFS, "public")

	if err := i18n.Init(localesSub); err != nil {
		log.Fatalf("i18n init: %v", err)
	}

	if err := server.Run(dataSub, publicSub); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		log.Fatalf("sub fs %q: %v", dir, err)
	}
	return sub
}
