package templatetargets

import (
	"context"
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"

	"github.com/srkn0/main/internal/templates/constants"
	"github.com/srkn0/main/internal/templates/ui/layouts"
	i18npkg "github.com/srkn0/main/pkg/i18n"
	"github.com/srkn0/main/pkg/util/render"
)

// writeTitleForHTMX emits a <title> element into the response body.
// HTMX picks up <title> in any swap response and sets document.title from
// it, so partial renders keep the tab title in sync without a full reload.
func writeTitleForHTMX(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprintf(w, "<title>%s</title>", html.EscapeString(i18npkg.PageTitle(ctx)))
	return err
}

var Main = render.TemplateTargets{
	"index": func(content templ.Component) templ.ComponentFunc {
		return func(ctx context.Context, w io.Writer) error {
			return layouts.Index(ctx).Render(templ.WithChildren(ctx, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				return layouts.MainPageLayout(ctx).Render(templ.WithChildren(ctx, content), w)
			})), w)
		}
	},
	constants.PageLayoutID: func(content templ.Component) templ.ComponentFunc {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			if err := writeTitleForHTMX(ctx, w); err != nil {
				return err
			}
			return layouts.MainPageLayout(ctx).Render(templ.WithChildren(ctx, content), w)
		})
	},
	constants.MainContentID: func(content templ.Component) templ.ComponentFunc {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			if err := writeTitleForHTMX(ctx, w); err != nil {
				return err
			}
			return content.Render(ctx, w)
		})
	},
}

var Print = render.TemplateTargets{
	"index": func(content templ.Component) templ.ComponentFunc {
		return func(ctx context.Context, w io.Writer) error {
			return layouts.Print(ctx).Render(templ.WithChildren(ctx, content), w)
		}
	},
}
