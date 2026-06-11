---
title: "Developing against a prod-like cluster with DevSpace and KinD"
description: "Local development against a real Kubernetes instead of mocks: a KinD cluster running the same Helm chart as production, a one-command bootstrap script, and the DevSpace hot-reload loop with file sync across several services. Why prod parity makes development more consistent."
tags: [kubernetes, devspace, kind, developer-experience, helm]
date: 2026-05-22
---

## The problem with local development

Local environments drift from production. A Docker Compose stack or mocks reproduce the database and HTTP, but not what the code actually meets in production: Kubernetes manifests, CRDs, RBAC, service discovery and ingress routing. Problems then surface only at deploy time.

For a Kubernetes operator the gap is especially wide. An operator reconciles custom resources through the Kubernetes API. Without a real API there is nothing to reconcile. The operator can only be developed meaningfully against a running cluster.

The setup described here comes from the [kubeplate](/projects/kubeplate) project and solves this by running the same thing locally as in production.

## A real cluster locally with KinD

KinD starts a full Kubernetes inside Docker containers. Into that cluster the same Helm chart is deployed that also runs in production: the same CRDs, the same RBAC, the same database, the same routing.

This is parity, not simulation. What works locally works in the cluster, because it is the same cluster.

```yaml
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080   # Traefik HTTP
        hostPort: 80
      - containerPort: 30443   # Traefik HTTPS
        hostPort: 443
```

The port mappings forward host ports 80 and 443 to the ingress in the cluster. Path-based routing is then the same locally as in production: `/api` goes to the server, `/` to the web UI.

## Bootstrap in one command

A script creates the cluster idempotently. It checks whether the cluster exists, creates it if needed, optionally installs Gateway API and Traefik, and rolls out the Helm chart.

```bash
./dev/bootstrap-kind.sh
```

Idempotency is the point here. The script can be run any number of times; an existing cluster is not rebuilt. Onboarding is then one command, not a page of instructions.

## The hot-reload loop with DevSpace

`devspace dev` replaces the application pods with development containers and connects the working directory to the cluster. In kubeplate three services run in parallel: the web UI, the server and the operator.

```yaml
dev:
  web:
    sync:
      - path: ./:/app
    ports:
      - port: "3000:3000"
    command: ["sh", "-c", "pnpm install && cd apps/web && next dev"]
```

Three mechanisms work together:

- File sync mirrors changed files into the pod immediately, without building a new image.
- Each stack hot-reloads: the web UI through the dev server, the server through a watch process, the operator through `air`, which rebuilds the Go binary on every change.
- Caches such as `node_modules` and the Go module cache live on persistent volumes and survive pod restarts.

A code change is then active in the cluster seconds later. The inner loop is as short as in purely local development, but runs against real Kubernetes.

## Developing against real infrastructure

The gain is consistency. The operator reconciles real custom resources, RBAC and service discovery behave like production, and ingress routing is the same. There is no longer a class of bugs that surfaces only after a deploy because something was missing locally.

Consistency holds across the team as well. Everyone develops against the same cluster definition and the same Helm chart. The pinned toolchain and the bootstrap script keep the environment identical on every machine. "It works on my machine" loses its meaning when "my machine" is the same everywhere.

## Workflow

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

Three commands from clone to live coding: install dependencies, bring up the cluster, start the hot-reload loop. After that, changes to the web UI, the server and the operator take effect in the cluster immediately.

## Summary

- Local mocks and compose stacks drift from production; for operators there is no basis without a real API
- KinD starts a real Kubernetes locally, into which the same Helm chart as production is deployed
- Port mappings bring real ingress routing onto the local host
- An idempotent bootstrap script reduces onboarding to one command
- `devspace dev` syncs files live and hot-reloads each service without an image rebuild
- Persistent caches keep the inner loop short
- Developing against real CRDs, RBAC and routing no longer defers a class of bugs to the deploy
- A shared cluster definition makes the environment consistent across the team
