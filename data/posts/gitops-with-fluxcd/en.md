---
title: "GitOps with FluxCD: patterns and antipatterns"
description: "Repository layout as an organizational decision with examples, bootstrap via the Flux CLI or the Flux Operator, and a set of antipatterns with the better pattern for each: secrets, CRD ordering, cluster duplication and version updates."
tags: [gitops, fluxcd, kubernetes, kustomize, homelab]
date: 2025-04-20
---

## GitOps with Flux

Flux polls a Git repository and continuously reconciles the cluster state against its contents. Git is the single source of truth; every change goes through a commit.

Two fields carry the approach. `prune: true` removes from the cluster whatever disappears from Git. `wait: true` marks a Kustomization ready only once its resources are healthy, and is the prerequisite for `dependsOn` to work.

The sections below are framed as patterns: a common antipattern, the better approach, and an example from a real homelab repo.

## Repository layout is a decision, not a standard

There is no correct Flux layout. The structure follows the organization: team size, number of clusters, and who owns which part. Three common layouts:

- Monorepo with one directory per cluster. Simple and clear for one team and a few clusters.
- Shared base with cluster overlays. Useful when several similar clusters run almost the same stack and differ only in variables.
- One repository per team or tenant. Fits large organizations with separate ownership and their own review rules.

The repo referenced here uses the shared base. App definitions live once under `apps/`; the clusters under `clusters/` reference them and carry only their selection and variables.

```text
kubernetes/
├── apps/          # HelmRelease and manifests per app
├── clusters/
│   ├── _base/     # shared Flux Kustomizations
│   └── home-01/   # per cluster: selection and variables
├── components/    # reusable Kustomize components
├── crds/          # CRDs, rolled out separately
└── repositories/  # HelmRepository and OCIRepository
```

## Bootstrap: two ways

Antipattern: establishing the cluster state imperatively via `kubectl apply` or through the UI. The state is then not recorded anywhere traceable.

Better: bootstrap Flux declaratively, so Flux manages itself from Git. There are two established ways.

The first is the Flux CLI. `flux bootstrap` commits the Flux components along with a `GitRepository` and a root Kustomization into the repo. Flux then manages its own installation too.

```bash
flux bootstrap github \
  --owner=example \
  --repository=homelab-k8s \
  --branch=main \
  --path=kubernetes/clusters/home-01
```

The second is the Flux Operator. A `FluxInstance` describes the distribution, the active controllers and the sync source as a custom resource. The Flux lifecycle then becomes declarative itself; upgrades and multiple instances get simpler.

Example from the repo: the `FluxInstance` sets the sync ref through a variable instead of hardcoding it.

```yaml
sync:
  kind: GitRepository
  url: https://github.com/example/homelab-k8s.git
  ref: ${FLUX_SYNC_REF}   # e.g. refs/heads/main or refs/tags/v1.4.0
  path: kubernetes/clusters/home-01/flux
```

The CLI is the most direct official entry point. The operator pays off once the Flux lifecycle itself should be versioned and pinned differently per cluster.

## Antipattern: plaintext secrets in the repo

Plaintext secrets in Git are the most common antipattern. Anyone with repo access reads them, and they stay in the history forever.

Better: store secrets encrypted, for example with SOPS and age, with Sealed Secrets, or via External Secrets from a vault. SOPS with age encrypts only the value fields, so the structure stays readable in the diff.

```yaml
# .sops.yaml
creation_rules:
  - path_regex: '.*\.sops\.ya?ml$'
    encrypted_regex: '^(data|stringData)$'
    age:
      - "<age-recipient>"
```

Flux decrypts at runtime with the private key, held as a secret in the cluster and not belonging in the repo. A Kustomize component enables decryption per Kustomization instead of setting it everywhere individually.

## Antipattern: CRDs and the app in the same release

When a CRD is managed by the same HelmRelease that also creates the matching custom resource, a race condition appears: the resource is applied before its CRD is registered.

Better: roll out CRDs separately. The HelmRelease carries `crds: Skip`, the CRDs run in their own Kustomization with `wait: true`, and the app depends on it via `dependsOn`.

```yaml
spec:
  dependsOn:
    - name: cert-manager-crds
  wait: true
```

This keeps the CRD lifecycle with Flux and guarantees the ordering.

## Antipattern: copying configuration per cluster

Maintaining several clusters by copy and paste leads to drift between them. A change has to be carried to several places and gets forgotten somewhere.

Better: a shared base plus variables. `postBuild.substitute` sets individual values, `substituteFrom` pulls them from a ConfigMap or a Secret.

```yaml
postBuild:
  substitute:
    CLUSTER_NAME: "home-01"
    CLUSTER_DOMAIN: "example.com"
  substituteFrom:
    - kind: ConfigMap
      name: cluster-vars
      optional: true
```

Example from the repo: the cluster-wide variables sit once in the root Kustomization and are inherited. Recurring fields like `sourceRef`, `interval`, `prune` and `wait` are set centrally by a Kustomize component, so every Kustomization inherits them instead of repeating them.

## Antipattern: maintaining versions by hand

Maintaining image tags and chart versions by hand is error-prone; `:latest` makes the state unreproducible.

Better: automate updates. Two approaches, depending on the flow you want.

Flux can update image tags itself. An `ImageRepository` scans the registry, an `ImagePolicy` selects the version, an `ImageUpdateAutomation` commits back to the repo. A marker comment links the field to the policy.

```yaml
image: ghcr.io/example/app:1.4.0 # {"$imagepolicy": "flux-system:app-policy"}
```

Alternatively an external tool like Renovate opens a pull request, a human reviews and merges. Example from the repo: chart versions carry a Renovate annotation and are updated through review.

```yaml
# renovate: datasource=helm depName=cert-manager registryUrl=https://charts.jetstack.io
CERT_MANAGER_VERSION: "v1.20.2"
```

Flux image automation fits when new app images should flow in automatically. The PR route fits when every change should go through review. Both are GitOps-conformant, because in the end the change is a commit.

## Summary

- Flux reconciles the cluster against Git; `prune` removes drift, `wait` enables `dependsOn`
- There is no correct layout; monorepo, shared base or repo per team follow the organization's needs
- Bootstrap declaratively, either via `flux bootstrap` or through the Flux Operator with a `FluxInstance`
- Pin the sync ref per cluster through a variable like `FLUX_SYNC_REF`
- Store secrets encrypted, for example with SOPS and age, never in plaintext
- Roll out CRDs separately with `wait: true`, attach apps via `dependsOn`
- Avoid cluster duplication through a shared base plus `postBuild` variables and components
- Update versions automatically, via Flux image automation or via a Renovate PR
