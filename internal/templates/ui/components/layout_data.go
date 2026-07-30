package components

import (
	"context"
	"net/url"
	"strings"

	"github.com/srkn0/main/internal/content"
)

type ExternalLink struct {
	Label string
	URL   string
}

type LayoutViewModel struct {
	CurrentPath   string
	LatestWriting []content.PostSummary
	ProjectLinks  []content.ProjectSummary
	ExternalLinks []ExternalLink
}

type layoutDataKey struct{}

func WithLayoutData(ctx context.Context, vm LayoutViewModel) context.Context {
	return context.WithValue(ctx, layoutDataKey{}, vm)
}

func LayoutData(ctx context.Context) LayoutViewModel {
	vm, _ := ctx.Value(layoutDataKey{}).(LayoutViewModel)
	if vm.CurrentPath == "" {
		vm.CurrentPath = "/"
	}
	return vm
}

func ActivePath(currentPath, href string) bool {
	if href == "/" {
		return currentPath == "/"
	}
	return currentPath == href || strings.HasPrefix(currentPath, href+"/")
}

func NavLinkClass(currentPath, href string) string {
	base := "group flex items-center gap-2 rounded-md px-2 py-1.5 text-sm leading-none outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
	if ActivePath(currentPath, href) {
		return base + " bg-accent text-foreground font-medium"
	}
	return base + " text-muted-foreground hover:bg-accent/70 hover:text-foreground"
}

func CompactLinkClass(currentPath, href string) string {
	if ActivePath(currentPath, href) {
		return "block rounded-md px-2 py-1.5 text-sm font-medium text-foreground bg-accent outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
	}
	return "block rounded-md px-2 py-1.5 text-sm text-muted-foreground hover:text-foreground hover:bg-accent/70 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
}

func LangURL(lang string) string {
	values := url.Values{}
	values.Set("set", lang)
	return "/lang?" + values.Encode()
}

func LangClass(current, lang string) string {
	base := "inline-flex h-8 min-w-8 items-center justify-center rounded-md px-2 text-xs font-medium outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
	if current == lang {
		return base + " bg-accent text-foreground"
	}
	return base + " text-muted-foreground hover:bg-accent/70 hover:text-foreground"
}
