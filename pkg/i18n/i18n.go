package i18n

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type contextKey struct{}

var localeKey = contextKey{}
var bundle *i18n.Bundle

func Init(fsys fs.FS) {
	bundle = i18n.NewBundle(language.German)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	dir, err := fs.ReadDir(fsys, ".")
	if err != nil {
		log.Fatalf("failed to read locales directory: %v", err)
	}
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			log.Printf("failed to read message file %s: %v", entry.Name(), err)
			continue
		}
		if _, err := bundle.ParseMessageFileBytes(data, entry.Name()); err != nil {
			log.Printf("failed to parse message file %s: %v", entry.Name(), err)
		}
	}
}

func WithLocale(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, localeKey, lang)
}

func GetLocale(ctx context.Context) string {
	if v, ok := ctx.Value(localeKey).(string); ok {
		return v
	}
	return "de"
}

func T(ctx context.Context, key string) string {
	localizer := i18n.NewLocalizer(bundle, GetLocale(ctx))
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: key})
	if err != nil {
		return key
	}
	return msg
}

func DateFormat(ctx context.Context) string {
	if GetLocale(ctx) == "en" {
		return "Jan 02, 2006"
	}
	return "02.01.2006"
}
