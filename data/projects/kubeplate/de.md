---
title: "kubeplate"
description: "Ein GitHub-Template für full-stack Kubernetes-Operator-Apps: Go-Operator, Hono/tRPC-Backend und Next.js-Web rund um eine Demo-CRD, mit Dockerfiles, Helm-Umbrella-Chart, GHCR/GitHub-Actions-CI und DevSpace-Hot-Reload gegen KinD."
tags: [kubernetes, operator, go, devspace, kind, helm, typescript]
date: 2026-05-20
category: template
featured: 3
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
- Container: Dockerfiles für Operator, Server, Web und Docs
- Packaging: ein Helm-Umbrella-Chart über Operator, Server, Web und Postgres
- Monorepo: pnpm-Workspaces und Turborepo, formatiert über Biome
- CI: GitHub Actions mit affected-only-Image-Builds, Helm-Lint/Render, CRD- und RBAC-Sync-Checks, GHCR-Push und OCI-Chart-Publish

Der Operator reconciled die `WebApp`-Resource in ein Deployment und einen Service. Server und Web verwalten dieselben Resources über die Kubernetes-API, sodass die Oberfläche zeigt, was der Operator tut.

Die Lieferkette ist Teil des Templates: Pull Requests verifizieren die betroffenen Docker-Images ohne Push, `main` pusht die geänderten Images nach GHCR, und Release-Tags bauen alle Images mit Semver-Tags plus Helm-Chart. Dadurch ist das Repo nicht nur ein Code-Skelett, sondern enthält auch den Weg vom Fork bis zu versionierten Artefakten.

kubeplate ist nicht als primäre Plattformlösung für große Organisationen gedacht. Dort gibt es oft schon eigene Plattformteams und Abstraktionen wie Cloud Foundry, Crossplane, interne Golden Paths oder Self-Service-Portale. Der Vorteil liegt hier woanders: Geschwindigkeit, Unabhängigkeit und ein lokaler, isolierter Ops-Lifecycle.

Dockerfiles, Helm-Chart, Bootstrap, Gateway-Routing und CI/CD stecken direkt im Template, damit ein kleines Team eine Kubernetes-native Idee in einem echten Cluster validieren kann, ohne auf eine zentrale Dev-Umgebung oder einen dauerhaft betriebenen Sandbox-Server angewiesen zu sein. KinD läuft lokal, DevSpace macht Hot-Reload direkt ins Cluster, und der Feedback-Loop ist dadurch sehr schnell und reibungsarm. Jeder Entwickler kann sein eigenes Cluster kaputtmachen, ohne anderen in die Quere zu kommen. Für Operator-Entwicklung und frühe Greenfield-Ideen ist genau dieser kurze Loop der eigentliche Mehrwert.

Wenn sich eine Idee bewährt, können die wiederverwendbaren Teile später immer noch in eine sauber versionierte Plattform-Toolchain wandern. kubeplate hält die erste Iteration bewusst klein.

## Developer Experience

Lokal läuft ein KinD-Cluster, in das das Helm-Chart deployt wird, inklusive CRDs, RBAC, Postgres und Routing. Damit verhält sich die lokale Umgebung wie Produktion, statt sie zu simulieren.

DevSpace fährt drei Dev-Container parallel: Web, Server und Operator. Dateien werden live in die Pods synchronisiert, jeder Stack lädt heiß neu, und Caches bleiben über Pod-Neustarts erhalten. Eine Codeänderung ist Sekunden später im Cluster aktiv.

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

Ein Bootstrap-Skript legt das Cluster idempotent an, optional mit Gateway API und Traefik für produktionsnahes Routing. Standardmäßig läuft das über `kubeplate.localhost`, was ohne DNS-Eintrag auf `127.0.0.1` auflöst. Für eigene Domains wird der Host im Bootstrap per `APP_HOST` oder im Helm-Chart über `web.ingress.host` gesetzt; bei TLS gehört dazu auch `web.ingress.tls` und eine passende `betterAuthUrl`. Bei lokalen Domains muss der DNS-Eintrag auf die lokale Host- oder Ingress-IP zeigen, und Router mit DNS-Rebindschutz müssen die Domain erlauben, sonst wird die private IP-Antwort oft still überschrieben.

Der Einstieg ist drei Befehle lang.

Mehr zum Dev-Loop und zur Prod-Parität im Artikel [Entwicklung gegen ein prod-nahes Cluster mit DevSpace und KinD](/blog/local-development-with-devspace-and-kind).
