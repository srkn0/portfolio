package i18n

import "context"

type contextKey struct{}

var localeKey = contextKey{}

func WithLocale(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, localeKey, lang)
}

func GetLocale(ctx context.Context) string {
	if v, ok := ctx.Value(localeKey).(string); ok {
		return v
	}
	return "de"
}

func T(_ context.Context, key string) string { return key }
