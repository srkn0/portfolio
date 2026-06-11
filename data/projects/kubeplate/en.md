---
title: "kubeplate"
description: "A GitHub template for full-stack Kubernetes operator apps: a Go operator, a Hono/tRPC backend and a Next.js web UI around a demo CRD, bundled in a Helm umbrella chart, with a DevSpace hot-reload dev loop against a KinD cluster and GitHub Actions CI."
tags: [kubernetes, operator, go, devspace, kind, helm, typescript]
date: 2026-05-20
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
- Packaging: a Helm umbrella chart over operator, server, web and Postgres
- Monorepo: pnpm workspaces and Turborepo, formatted with Biome
- CI: GitHub Actions with affected-only builds, Helm lint, a CRD sync check and a GHCR push

The operator reconciles the `WebApp` resource into a Deployment and a Service. Server and web manage the same resources through the Kubernetes API, so the UI shows what the operator does.

## Developer experience

Locally a KinD cluster runs, into which the Helm chart is deployed, including CRDs, RBAC, Postgres and routing. The local environment thus behaves like production instead of simulating it.

DevSpace runs three dev containers in parallel: web, server and operator. Files are synced live into the pods, each stack hot-reloads, and caches survive pod restarts. A code change is active in the cluster seconds later.

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

A bootstrap script creates the cluster idempotently, optionally with Gateway API and Traefik for production-like routing. Onboarding is three commands long.

More on the dev loop and prod parity in the article [Developing against a prod-like cluster with DevSpace and KinD](/blog/local-development-with-devspace-and-kind).
