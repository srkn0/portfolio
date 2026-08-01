---
title: "devhub"
description: "AI-assisted Platform-Lab für Kubernetes-Self-Service: Teams können Datenbanken, Services und IDE-Workspaces aus einer Weboberfläche deployen, während Operator, API und GitOps-Setup die Kubernetes-Seite kapseln."
tags: [kubernetes, platform-engineering, operator, gitops, tanstack, hono, trpc]
date: 2026-08-01
category: platform
featured: 4
repo: https://github.com/srkn0/devhub
---

## Überblick

devhub ist ein experimentelles Self-Service-Portal für Kubernetes. Entwickler bekommen eine Weboberfläche für Umgebungen, Datenbanken, Services und In-Cluster-Workspaces, ohne selbst mit kubeconfig, Helm oder kubectl arbeiten zu müssen.

Das Projekt verbindet Platform-Engineering-Flows, Operator-Pattern, Multi-Tenancy und moderne Full-Stack-Entwicklung in einem vertikalen Schnitt durch eine interne Developer Platform.

## Stack & Architektur

- Web: Next.js, React, TanStack Query und TailwindCSS
- API: Hono, tRPC, Better-Auth und Kubernetes-Client
- Operator: Go mit controller-runtime
- Packaging: Helm-Chart mit Operator, Server und Web
- Local Dev: KinD, DevSpace und Gateway API
- GitOps: Deployment-Pfad über mein Homelab-Cluster

Die UI spricht nicht direkt mit Kubernetes. Requests laufen über API und Operator, Status wird aus dem Cluster zurückgespiegelt. Dadurch lässt sich die Plattform-Idee produktionsnah testen, ohne die Komplexität einer echten zentralen Plattformorganisation zu brauchen.

## Entwicklungsmodell

AI ist Teil der Toolchain: Hypothese bauen, vertikale Scheibe deployen, gegen ein echtes Cluster testen, dann gezielt härten. Dadurch entstehen Platform-Prototypen schnell genug, um Architekturentscheidungen früh an echten Kubernetes-Ressourcen zu prüfen.
