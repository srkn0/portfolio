---
title: "Portfolio mit Go, HTMX und Templ"
description: "Persönliches Portfolio und Blog. Go im Backend, HTMX im Frontend, Templ als Template-Engine, Markdown als Content-Store. Deployment als Single Binary mit allem embedded."
tags: [go, htmx, templ, tailwind, markdown]
date: 2026-05-10
category: lab
repo: https://github.com/srkn0/portfolio
---

## Überblick

Persönliches Portfolio mit Blog, Projektliste und CV. Die Seite läuft als einzelnes Go-Binary. Content liegt als Markdown im Repository, wird beim Build embedded und beim Start in den Speicher geladen.

## Credits

Als Ausgangspunkt diente [go-htmx-starter](https://github.com/carsonkrueger/go-htmx-starter) von Carson Krueger. Vor allem das Zusammenspiel von Templ und HTMX kommt von dort. Seitdem ist viel ergänzt und umgebaut worden.

## Stack

- Go 1.26.2
- Chi als HTTP-Router
- Templ für typsichere Templates, compiliert zu Go-Code
- TemplUI als Komponenten-Bibliothek auf Templ-Basis
- HTMX und kleines Vanilla-JavaScript fürs Frontend-Verhalten
- Tailwind CSS v4
- Goldmark als Markdown-Parser mit GFM, Footnotes und Frontmatter
- go-i18n für Deutsch und Englisch
- Air für Hot-Reload im Dev-Modus
- mise tasks für Codegenerierung und Dev-Workflow

Im Production-Build landen App-Code, Markdown-Dateien sowie lokale CSS-, Font- und JS-Assets in einem Binary. Content und lokale Assets liegen in `embed.FS`, ohne externes Filesystem zur Laufzeit.

## Architektur

Drei Bereiche.

**Content.** Blogposts, Projekte und CV liegen als Markdown-Dateien unter `data/`. Beim Start parst Goldmark die Files und legt sie in In-Memory-Stores ab. Schlüssel ist der Slug, darunter eine Map pro Locale.

**Templates.** Die UI liegt in `internal/templates/ui/`. Layouts, Pages, Components. Templ generiert daraus typsichere Go-Funktionen. Template-Fehler werden beim Build sichtbar, nicht erst zur Laufzeit.

**Server.** Chi-Router mit Middleware-Stack (Logger, Recoverer, Metrics, i18n-Cookie-Lookup). Handler holen Content aus dem In-Memory-Store, rendern das passende Template und antworten je nach HTMX-Header mit voller Seite oder Fragment.

## Mehrsprachigkeit

Dateinamen-Konvention `{slug}/de.md` und `{slug}/en.md` (Posts und Projects). Die Locale wird aus einem Cookie, einem Query-Parameter oder dem Accept-Language-Header bestimmt. Fallback ist Deutsch.

Statische Strings werden über go-i18n und JSON-Files in `locales/` übersetzt.

## HTMX-Layout

Beim ersten Request einer Seite liefert der Server das volle HTML mit Layout. Bei einem internen Link sendet HTMX einen Request mit `HX-Request: true`, und der Server antwortet nur mit dem Content-Fragment. Navigation und JS-Status bleiben erhalten, ohne SPA-Framework.

Die Layout-Auswahl passiert in `pkg/util/render`, anhand des HTMX-Headers.

## Dark Mode

CSS plus kleines Vanilla-JavaScript. Toggle setzt eine Klasse am `html`-Tag und schreibt den Wert in `localStorage`. Beim ersten Laden wird der gespeicherte Wert oder die System-Preference ausgelesen, bevor das CSS angewendet wird, damit kein Flash entsteht.
