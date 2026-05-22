package templatetargets

import (
	"context"
	"io"

	"github.com/a-h/templ"

	"github.com/srkn0/main/internal/templates/constants"
	"github.com/srkn0/main/internal/templates/ui/layouts"
	"github.com/srkn0/main/pkg/util/render"
)

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
			return layouts.MainPageLayout(ctx).Render(templ.WithChildren(ctx, content), w)
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
