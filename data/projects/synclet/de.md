---
title: "synclet"
description: "AI-assisted Django/Vue-Referenzanwendung zur praktischen Vertiefung von Kontaktsynchronisation, typisierten APIs, Validierung, Tests und Delivery-Workflows."
tags: [python, django, vue, postgresql, uv, turborepo, biome]
date: 2026-08-03
category: lab
repo: https://github.com/srkn0/synclet
---

## Überblick

synclet ist eine AI-gestützte Django/Vue-Referenzanwendung, die einen kleinen, realistischen Use Case rund um Kontaktsynchronisation umsetzt. Kontakte können verwaltet und über CSV-Dateien importiert werden; Synchronisationsläufe speichern Status, Zähler, Duplikatbehandlung und Fehler.

Das Projekt entstand, um praktische Erfahrung mit Python, Django, Vue und modernen Full-Stack-Delivery-Workflows aufzubauen. Der Umfang bleibt bewusst begrenzt: synclet ist keine vollständige SaaS-Plattform und implementiert keine echten Google-, Microsoft- oder CRM-Integrationen.

## Stack & Architektur

- Django und Django REST Framework als Backend
- Vue 3, Vite und Nuxt UI als Weboberfläche
- PostgreSQL als persistenter Datenspeicher
- CSV-Import hinter einer `SyncProvider`-Abstraktion
- OpenAPI als API-Vertrag mit generierten TypeScript-Typen
- pnpm Workspaces und Turborepo für das Monorepo
- `uv` für Python-Abhängigkeiten und Lockfiles
- mise zur Versionierung der Toolchain
- Biome für JavaScript-, TypeScript-, Vue- und JSON-Linting/Formatting
- Taskfile für reproduzierbare Entwicklungsbefehle

Das Backend bleibt ein modularer Django-Monolith. Die Vue-SPA kommuniziert über eine typisierte REST-API mit Django. Importverarbeitung, Validierung, Benutzerisolation und Datenbanktransaktionen liegen im Backend; das Frontend bildet die Zustände und Ergebnisse der Synchronisationsläufe ab.

## Qualität & Sicherheit

Der Referenz-Use-Case umfasst authentifizierte, benutzerisolierte Contact-CRUDs, Suche, Pagination, CSRF-Schutz und einen atomaren CSV-Import mit Größen- und Zeilenlimits. Tests decken unter anderem Authentifizierung, Isolation, CSV-Validierung, Duplikate, Upserts, Rollback-Verhalten und Fehlerfälle ab.

Zusätzlich gibt es OpenAPI-/TypeScript-Vertragsprüfungen, Vue- und API-Tests, einen Playwright-End-to-End-Test, Docker-Healthchecks, strukturierte Request-Logs, Prometheus-kompatible Metriken und eine GitHub-Actions-CI.

## Bewusste Grenzen

Der CSV-Provider arbeitet synchron und ist die einzige implementierte Quelle. Provider für Google Contacts, Microsoft Graph oder CRM-Systeme sind nur als mögliche Erweiterung der Architektur vorgesehen. Eine persistente Queue, Multi-Tenancy, Rollen, Audit-Logs und Provider-Credentials gehören zu den dokumentierten nächsten Ausbaustufen.
