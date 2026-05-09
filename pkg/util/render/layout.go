package render

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/a-h/templ"

	"github.com/srkn0/main/pkg/util"
)

const indexTarget = "index"

type TargetFunc func(content templ.Component) templ.ComponentFunc
type TemplateTargets map[string]TargetFunc

var ErrMissingIndexTarget = errors.New("template targets is missing required index entry")

func Layout(ctx context.Context, req *http.Request, w io.Writer, targets TemplateTargets, page templ.Component) error {
	if !util.IsHxRequest(req) {
		index, ok := targets[indexTarget]
		if !ok {
			return ErrMissingIndexTarget
		}
		return index(page).Render(ctx, w)
	}

	if target, ok := targets[req.Header.Get("HX-Target")]; ok {
		return target(page).Render(ctx, w)
	}

	return page.Render(ctx, w)
}
