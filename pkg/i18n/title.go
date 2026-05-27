package i18n

import "context"

type ctxKeyPageTitle struct{}

// WithPageTitle attaches a per-page title to the context. The Index layout
// reads it via PageTitle to render the <title> tag.
func WithPageTitle(ctx context.Context, title string) context.Context {
	return context.WithValue(ctx, ctxKeyPageTitle{}, title)
}

// PageTitle returns the page title set via WithPageTitle, or "Portfolio"
// as a fallback when no handler set one.
func PageTitle(ctx context.Context) string {
	if t, ok := ctx.Value(ctxKeyPageTitle{}).(string); ok && t != "" {
		return t
	}
	return "Portfolio"
}
