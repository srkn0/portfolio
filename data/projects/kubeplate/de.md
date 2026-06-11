---
title: "kubeplate"
description: "Ein GitHub-Template für full-stack Kubernetes-Operator-Apps: Go-Operator, Hono/tRPC-Backend und Next.js-Web rund um eine Demo-CRD, gebündelt in einem Helm-Umbrella-Chart, mit DevSpace-Hot-Reload-Dev-Loop gegen ein KinD-Cluster und GitHub-Actions-CI."
tags: [kubernetes, operator, go, devspace, kind, helm, typescript]
date: 2026-05-20
repo: https://github.com/srkn0/kubeplate
---

## Überblick

kubeplate ist ein Template zum Bauen vollständiger Kubernetes-Operator-Anwendungen. Es bündelt einen Operator, ein Backend, eine Weboberfläche, eine Datenbank und ein Helm-Chart rund um eine Demo-`WebApp`-Custom-Resource. Forken, die Demo-CRD durch die eigene ersetzen, ausliefern.

Der Reiz liegt im Dev-Loop: Entwickelt wird gegen ein echtes Kubernetes, nicht gegen einen Mock.

## Stack & Architektur

- Operator: Go mit controller-runtime und kubebuilder, getestet über envtest
- API: Hono und tRPC mit Better-Auth und dem Kubernetes-Client
- Web: Next.js und React mit TailwindCSS und shadcn/ui
- Datenbank: PostgreSQL mit Prisma
- Packaging: ein Helm-Umbrella-Chart über Operator, Server, Web und Postgres
- Monorepo: pnpm-Workspaces und Turborepo, formatiert über Biome
- CI: GitHub Actions mit affected-only-Builds, Helm-Lint, CRD-Sync-Check und GHCR-Push

Der Operator reconciled die `WebApp`-Resource in ein Deployment und einen Service. Server und Web verwalten dieselben Resources über die Kubernetes-API, sodass die Oberfläche zeigt, was der Operator tut.

## Developer Experience

Lokal läuft ein KinD-Cluster, in das das Helm-Chart deployt wird, inklusive CRDs, RBAC, Postgres und Routing. Damit verhält sich die lokale Umgebung wie Produktion, statt sie zu simulieren.

DevSpace fährt drei Dev-Container parallel: Web, Server und Operator. Dateien werden live in die Pods synchronisiert, jeder Stack lädt heiß neu, und Caches bleiben über Pod-Neustarts erhalten. Eine Codeänderung ist Sekunden später im Cluster aktiv.

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

Ein Bootstrap-Skript legt das Cluster idempotent an, optional mit Gateway API und Traefik für produktionsnahes Routing. Der Einstieg ist drei Befehle lang.

Mehr zum Dev-Loop und zur Prod-Parität im Artikel [Entwicklung gegen ein prod-nahes Cluster mit DevSpace und KinD](/blog/local-development-with-devspace-and-kind).
