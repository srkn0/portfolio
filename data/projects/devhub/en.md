---
title: "devhub"
description: "Platform lab for Kubernetes self-service: teams can deploy databases, services, and IDE workspaces from a web UI while the operator, API, and GitOps setup hide the Kubernetes details."
tags: [kubernetes, platform-engineering, operator, gitops, tanstack, hono, trpc]
date: 2026-08-01
category: platform
featured: 4
repo: https://github.com/srkn0/devhub
---

## Overview

devhub is an experimental self-service portal for Kubernetes. Developers get a web UI for environments, databases, services, and in-cluster workspaces without needing direct kubeconfig, Helm, or kubectl access.

The project connects platform-engineering flows, operator patterns, multi-tenancy, and modern full-stack development into a compact prototype of an internal developer platform.

## Stack & Architecture

- Web: Next.js, React, TanStack Query, and TailwindCSS
- API: Hono, tRPC, Better-Auth, and the Kubernetes client
- Operator: Go with controller-runtime
- Packaging: Helm chart with operator, server, and web
- Local dev: KinD, DevSpace, and Gateway API
- GitOps: deployment path through my homelab cluster

The UI does not talk to Kubernetes directly. Requests flow through the API and operator, and status is reflected back from the cluster. That makes the platform idea testable in a production-like setup without needing the full weight of a central platform organization.

## Note

The project was created with AI support.
