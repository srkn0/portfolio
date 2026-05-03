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
	localesSub, _ := fs.Sub(localesFS, "locales")
	dataSub, _ := fs.Sub(dataFS, "data")
	publicSub, _ := fs.Sub(publicFS, "public")

	i18n.Init(localesSub)

	if err := server.Run(dataSub, publicSub); err != nil {
		log.Fatal(err)
	}
}
