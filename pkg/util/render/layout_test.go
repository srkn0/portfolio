package render_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"

	"github.com/srkn0/main/pkg/util/render"
)

// component returns a templ.Component that writes the given marker so we can
// assert which template path was taken.
func component(marker string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, marker)
		return err
	})
}

// wrapper produces a TargetFunc that wraps the inner component in <prefix>...</suffix>
func wrapper(prefix, suffix string) render.TargetFunc {
	return func(c templ.Component) templ.ComponentFunc {
		return func(ctx context.Context, w io.Writer) error {
			if _, err := io.WriteString(w, prefix); err != nil {
				return err
			}
			if err := c.Render(ctx, w); err != nil {
				return err
			}
			_, err := io.WriteString(w, suffix)
			return err
		}
	}
}

func TestLayout_fullPageUsesIndexTarget(t *testing.T) {
	targets := render.TemplateTargets{
		"index": wrapper("<index>", "</index>"),
	}
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := render.Layout(req.Context(), req, &buf, targets, component("PAGE")); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	if got := buf.String(); got != "<index>PAGE</index>" {
		t.Errorf("output = %q, want index-wrapped page", got)
	}
}

func TestLayout_missingIndexTargetReturnsSentinel(t *testing.T) {
	targets := render.TemplateTargets{} // intentionally empty
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := render.Layout(req.Context(), req, io.Discard, targets, component("PAGE"))
	if !errors.Is(err, render.ErrMissingIndexTarget) {
		t.Errorf("err = %v, want ErrMissingIndexTarget", err)
	}
}

func TestLayout_hxRequestUsesNamedTarget(t *testing.T) {
	targets := render.TemplateTargets{
		"index":        wrapper("<index>", "</index>"),
		"main-content": wrapper("<main>", "</main>"),
	}
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "main-content")

	if err := render.Layout(req.Context(), req, &buf, targets, component("PAGE")); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	if got := buf.String(); got != "<main>PAGE</main>" {
		t.Errorf("output = %q, want main-wrapped page", got)
	}
}

func TestLayout_hxRequestUnknownTargetRendersPageBare(t *testing.T) {
	targets := render.TemplateTargets{
		"index": wrapper("<index>", "</index>"),
	}
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "does-not-exist")

	if err := render.Layout(req.Context(), req, &buf, targets, component("PAGE")); err != nil {
		t.Fatalf("Layout: %v", err)
	}
	if got := buf.String(); got != "PAGE" {
		t.Errorf("output = %q, want bare page", got)
	}
}
