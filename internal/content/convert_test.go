package content

import (
	"strings"
	"testing"
	"time"
)

func TestIsSupportedLocale(t *testing.T) {
	cases := map[string]bool{
		"de": true,
		"en": true,
		"fr": false,
		"":   false,
	}
	for in, want := range cases {
		if got := isSupportedLocale(in); got != want {
			t.Errorf("isSupportedLocale(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestConvertMarkdown_basic(t *testing.T) {
	src := []byte(`---
title: "Hello"
date: 2026-05-01
---

# Greeting

Hello, **world**.`)

	html, meta, err := convertMarkdown(src)
	if err != nil {
		t.Fatalf("convertMarkdown: %v", err)
	}
	if !strings.Contains(html, "<h1") {
		t.Errorf("output missing <h1>: %s", html)
	}
	if !strings.Contains(html, "<strong>world</strong>") {
		t.Errorf("output missing <strong>world</strong>: %s", html)
	}
	if meta["title"] != "Hello" {
		t.Errorf("meta[title] = %v, want Hello", meta["title"])
	}
}

func TestConvertMarkdown_autoHeadingID(t *testing.T) {
	html, _, err := convertMarkdown([]byte("## Hello World\n"))
	if err != nil {
		t.Fatalf("convertMarkdown: %v", err)
	}
	if !strings.Contains(html, `id="hello-world"`) {
		t.Errorf("expected auto heading ID, got: %s", html)
	}
}

func TestParseMeta_allFields(t *testing.T) {
	meta := map[string]any{
		"title":       "T",
		"description": "D",
		"tags":        []any{"go", "test"},
		"date":        "2026-05-01",
	}
	title, desc, tags, date, err := parseMeta(meta)
	if err != nil {
		t.Fatalf("parseMeta: %v", err)
	}
	if title != "T" {
		t.Errorf("title = %q", title)
	}
	if desc != "D" {
		t.Errorf("description = %q", desc)
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "test" {
		t.Errorf("tags = %v", tags)
	}
	want, _ := time.Parse("2006-01-02", "2026-05-01")
	if !date.Equal(want) {
		t.Errorf("date = %v, want %v", date, want)
	}
}

func TestParseMeta_dateAsTime(t *testing.T) {
	when := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, _, _, got, err := parseMeta(map[string]any{"date": when})
	if err != nil {
		t.Fatalf("parseMeta: %v", err)
	}
	if !got.Equal(when) {
		t.Errorf("date = %v, want %v", got, when)
	}
}

func TestParseMeta_invalidDateReturnsError(t *testing.T) {
	_, _, _, _, err := parseMeta(map[string]any{"date": "not-a-date"})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestParseMeta_ignoresNonStringTags(t *testing.T) {
	// YAML decoding lands in []any.
	// dropping non-strings instead of blowing up the whole post load.
	_, _, tags, _, err := parseMeta(map[string]any{
		"tags": []any{"go", 42, true, "test"},
	})
	if err != nil {
		t.Fatalf("parseMeta: %v", err)
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "test" {
		t.Errorf("tags = %v, want [go test]", tags)
	}
}
