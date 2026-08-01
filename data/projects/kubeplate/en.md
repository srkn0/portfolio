---
title: "kubeplate"
description: "A GitHub template for full-stack Kubernetes operator apps: a Go operator, a Hono/tRPC backend and a Next.js web UI around a demo CRD, with Dockerfiles, a Helm umbrella chart, GHCR/GitHub Actions CI and a DevSpace hot-reload loop against KinD."
tags: [kubernetes, operator, go, devspace, kind, helm, typescript]
date: 2026-05-20
category: template
featured: 3
repo: https://github.com/srkn0/kubeplate
---

## Overview

kubeplate is a template for building complete Kubernetes operator applications. It bundles an operator, a backend, a web UI, a database and a Helm chart around a demo `WebApp` custom resource. Fork it, replace the demo CRD with your own, ship.

The appeal is the dev loop: development happens against a real Kubernetes, not a mock.

## Stack & architecture

- Operator: Go with controller-runtime and kubebuilder, tested via envtest
- API: Hono and tRPC with Better-Auth and the Kubernetes client
- Web: Next.js and React with TailwindCSS and shadcn/ui
- Database: PostgreSQL with Prisma
- Containers: Dockerfiles for operator, server, web and docs
- Packaging: a Helm umbrella chart over operator, server, web and Postgres
- Monorepo: pnpm workspaces and Turborepo, formatted with Biome
- CI: GitHub Actions with affected-only image builds, Helm lint/render, CRD and RBAC sync checks, GHCR pushes and OCI chart publishing

The operator reconciles the `WebApp` resource into a Deployment and a Service. Server and web manage the same resources through the Kubernetes API, so the UI shows what the operator does.

The delivery path is part of the template: pull requests verify the affected Docker images without pushing, `main` pushes changed images to GHCR, and release tags build all images with semver tags plus the Helm chart. That makes the repo more than a code skeleton; it includes the route from fork to versioned artifacts.

kubeplate is not meant to be the primary platform solution for a large organization. Those environments often already have dedicated platform teams and abstractions such as Cloud Foundry, Crossplane, internal golden paths or self-service portals. The advantage here is different: speed, independence and a local, isolated ops lifecycle.

Dockerfiles, the Helm chart, bootstrap scripts, Gateway routing and CI/CD live directly in the template so a small team can validate a Kubernetes-native idea in a real cluster without depending on a central dev environment or a permanently operated sandbox server. KinD runs locally, DevSpace hot-reloads directly into the local cluster, and changes come back with very little friction. Every developer can break their own cluster without blocking anyone else. For operator development and early greenfield ideas, that short feedback loop is the real value.

If an idea proves itself, the reusable parts can still move into a properly versioned platform toolchain later. kubeplate deliberately keeps the first iteration small.

## Developer experience

Locally a KinD cluster runs, into which the Helm chart is deployed, including CRDs, RBAC, Postgres and routing. The local environment thus behaves like production instead of simulating it.

DevSpace runs three dev containers in parallel: web, server and operator. Files are synced live into the pods, each stack hot-reloads, and caches survive pod restarts. A code change is active in the cluster seconds later.

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

A bootstrap script creates the cluster idempotently, optionally with Gateway API and Traefik for production-like routing. By default this uses `kubeplate.localhost`, which resolves to `127.0.0.1` without a DNS entry. For custom domains, set the hostname via `APP_HOST` in the bootstrap flow or `web.ingress.host` in the Helm chart; for TLS, also set `web.ingress.tls` and a matching `betterAuthUrl`. Local-only domains need DNS pointing at the local host or ingress IP, and routers with DNS rebind protection may need an allowlist entry because they can silently replace private-IP DNS answers.

Onboarding is three commands long.

More on the dev loop and prod parity in the article [Developing against a prod-like cluster with DevSpace and KinD](/blog/local-development-with-devspace-and-kind).
