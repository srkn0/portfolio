package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const DefaultLocale = "de"

type contextKey struct{}

var localeKey = contextKey{}

var bundle *i18n.Bundle

func Init(fsys fs.FS) error {
	b := i18n.NewBundle(language.German)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return fmt.Errorf("reading message file %s: %w", entry.Name(), err)
		}
		if _, err := b.ParseMessageFileBytes(data, entry.Name()); err != nil {
			return fmt.Errorf("parsing message file %s: %w", entry.Name(), err)
		}
	}

	bundle = b
	return nil
}

func WithLocale(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, localeKey, lang)
}

func GetLocale(ctx context.Context) string {
	if v, ok := ctx.Value(localeKey).(string); ok {
		return v
	}
	return DefaultLocale
}

func T(ctx context.Context, key string) string {
	return Tc(ctx, key, nil)
}

func Tc(ctx context.Context, key string, data map[string]any) string {
	localizer := i18n.NewLocalizer(bundle, GetLocale(ctx))
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return msg
}

func DateFormat(ctx context.Context) string {
	switch GetLocale(ctx) {
	case "en":
		return "Jan 02, 2006"
	default:
		return "02.01.2006"
	}
}
