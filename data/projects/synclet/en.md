---
title: "synclet"
description: "Django/Vue reference application for exploring contact synchronization, typed APIs, validation, testing, and delivery workflows."
tags: [python, django, vue, postgresql, uv, turborepo, biome]
date: 2026-08-03
category: lab
repo: https://github.com/srkn0/synclet
---

## Overview

synclet is a Django/Vue reference application built around a small but realistic contact synchronization use case. Contacts can be managed and imported from CSV files; synchronization runs persist status, counters, duplicate handling, and errors.

The project was created with AI support to build practical experience with Python, Django, Vue, and modern full-stack delivery workflows. Its scope is deliberately limited: synclet is not a complete SaaS platform and does not implement real Google, Microsoft, or CRM integrations.

## Stack & architecture

- Django and Django REST Framework as the backend
- Vue 3, Vite, and Nuxt UI for the web application
- PostgreSQL as the persistent datastore
- CSV import behind a `SyncProvider` abstraction
- OpenAPI as the API contract with generated TypeScript types
- pnpm workspaces and Turborepo for the monorepo
- `uv` for Python dependencies and lockfiles
- mise for toolchain version management
- Biome for JavaScript, TypeScript, Vue, and JSON linting/formatting
- Taskfile for reproducible development commands

The backend remains a modular Django monolith. The Vue SPA communicates with Django through a typed REST API. Import processing, validation, user isolation, and database transactions live in the backend; the frontend represents synchronization states and results.

## Quality & security

The reference use case includes authenticated, user-isolated contact CRUD, search, pagination, CSRF protection, and an atomic CSV import with file-size and row limits. Tests cover authentication, isolation, CSV validation, duplicates, upserts, rollback behavior, and error cases.

The repository also includes OpenAPI/TypeScript contract checks, Vue and API tests, a Playwright end-to-end test, Docker healthchecks, structured request logs, Prometheus-compatible metrics, and GitHub Actions CI.

## Deliberate boundaries

CSV processing is synchronous and CSV is the only implemented provider. Google Contacts, Microsoft Graph, and CRM providers are architectural extension points rather than fake implementations. A durable queue, multi-tenancy, roles, audit logs, and provider credentials are documented as possible next steps.
