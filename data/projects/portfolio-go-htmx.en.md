---
title: "Portfolio with Go, HTMX and Templ"
description: "Personal portfolio and blog. Go on the backend, HTMX on the frontend, Templ as the template engine, Markdown as the content store. Ships as a single binary with everything embedded."
tags: [go, htmx, templ, tailwind, markdown]
date: 2026-05-10
repo: https://github.com/srkn0/portfolio
---

## Overview

Personal portfolio with a blog, project listing and CV. The site runs as a single Go binary. Content lives as Markdown in the repository, gets embedded at build time and loaded into memory at startup.

## Credits

The initial scaffold was [go-htmx-starter](https://github.com/carsonkrueger/go-htmx-starter) by Carson Krueger. The Templ + HTMX interplay came from there. A lot has been added and reshaped since.

## Stack

- Go 1.24+
- Chi as the HTTP router
- Templ for type-safe templates, compiled to Go code
- TemplUI as a component library on top of Templ
- HTMX and Hyperscript for frontend behavior
- Tailwind CSS v4
- Goldmark as the Markdown parser with GFM, footnotes, frontmatter
- go-i18n for German and English
- Air for hot reload in dev
- Taskfile as the task runner

In production everything compiles into a single binary. Markdown files, CSS, fonts and JS live in `embed.FS`, no external filesystem at runtime.

## Architecture

Three areas.

**Content.** Blog posts, projects and CV sit as Markdown files under `data/`. At startup Goldmark parses them into in-memory stores. Key is the slug, then a map per locale.

**Templates.** The UI lives in `internal/templates/ui/`. Layouts, pages, components. Templ generates type-safe Go functions. Template errors surface at build time, not in the browser.

**Server.** Chi router with a middleware stack (logger, recoverer, i18n cookie lookup). Handlers pull content from the in-memory store, render the matching template and reply with a full page or just a fragment depending on the HTMX header.

## i18n

Filename convention is `{slug}/de.md` and `{slug}/en.md` for posts and projects. The locale is resolved from a cookie, a query parameter, or the Accept-Language header. Default is German.

Static strings go through go-i18n and JSON files in `locales/`.

## HTMX layout

The first request to a page returns the full HTML with the layout. For internal links HTMX sends a request with `HX-Request: true`, and the server responds with just the content fragment. Header, footer and JS state stay alive, without an SPA framework.

Layout selection happens in `pkg/util/render`, based on the HTMX header.

## Dark mode

CSS plus a bit of Hyperscript. The toggle sets a class on the `html` tag and writes to `localStorage`. On first load the stored value or system preference is read before the CSS applies, so there is no flash.

## Planned

- RSS feed for the blog
- HTMX-driven search suggestions
- Optional static export
