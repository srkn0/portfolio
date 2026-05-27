package i18n_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/srkn0/main/pkg/i18n"
)

func localeFS(files map[string]string) fstest.MapFS {
	m := fstest.MapFS{}
	for name, body := range files {
		m[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

func TestInit_loadsMessages(t *testing.T) {
	fs := localeFS(map[string]string{
		"de.json": `{"hello":"Hallo"}`,
		"en.json": `{"hello":"Hi"}`,
	})
	if err := i18n.Init(fs); err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx := i18n.WithLocale(context.Background(), "de")
	if got := i18n.T(ctx, "hello"); got != "Hallo" {
		t.Errorf("T(de, hello) = %q, want Hallo", got)
	}
	ctx = i18n.WithLocale(context.Background(), "en")
	if got := i18n.T(ctx, "hello"); got != "Hi" {
		t.Errorf("T(en, hello) = %q, want Hi", got)
	}
}

func TestInit_missingKeyReturnsKey(t *testing.T) {
	// When the translation is missing we want the lookup key back so the
	// page at least shows something instead of an empty string.
	fs := localeFS(map[string]string{"de.json": `{"hello":"Hallo"}`})
	if err := i18n.Init(fs); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := i18n.T(context.Background(), "missing"); got != "missing" {
		t.Errorf("T(missing) = %q, want fallback to key", got)
	}
}

func TestInit_invalidJSONReturnsError(t *testing.T) {
	fs := localeFS(map[string]string{"de.json": `{not valid json`})
	if err := i18n.Init(fs); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestGetLocale_defaultsWhenAbsent(t *testing.T) {
	if got := i18n.GetLocale(context.Background()); got != i18n.DefaultLocale {
		t.Errorf("GetLocale = %q, want %q", got, i18n.DefaultLocale)
	}
}

func TestTc_substitutesTemplateData(t *testing.T) {
	fs := localeFS(map[string]string{
		"de.json": `{"greeting":"Hallo {{.Name}}"}`,
	})
	if err := i18n.Init(fs); err != nil {
		t.Fatalf("Init: %v", err)
	}
	ctx := i18n.WithLocale(context.Background(), "de")
	got := i18n.Tc(ctx, "greeting", map[string]any{"Name": "Max"})
	if got != "Hallo Max" {
		t.Errorf("Tc = %q, want Hallo Max", got)
	}
}
