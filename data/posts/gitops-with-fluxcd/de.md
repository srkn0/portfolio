---
title: "GitOps mit FluxCD: Patterns und Antipatterns"
description: "Repository-Layout als Organisationsentscheidung mit Beispielen, Bootstrap über die Flux-CLI oder den Flux Operator, und eine Reihe von Antipatterns mit dem jeweils besseren Muster: Secrets, CRD-Reihenfolge, Cluster-Duplikation und Versions-Updates."
tags: [gitops, fluxcd, kubernetes, kustomize, homelab]
date: 2025-04-20
---

## GitOps mit Flux

Flux pollt ein Git-Repository und gleicht den Cluster-Zustand kontinuierlich gegen dessen Inhalt ab. Git ist die Single Source of Truth; jede Änderung läuft über einen Commit.

Zwei Felder tragen den Ansatz. `prune: true` entfernt aus dem Cluster, was aus Git verschwindet. `wait: true` markiert eine Kustomization erst als bereit, wenn ihre Ressourcen gesund sind, und ist die Voraussetzung dafür, dass `dependsOn` greift.

Die folgenden Abschnitte sind als Muster formuliert: ein verbreitetes Antipattern, das bessere Vorgehen und ein Beispiel aus einem realen Homelab-Repo.

## Repository-Layout ist eine Entscheidung, kein Standard

Es gibt kein richtiges Flux-Layout. Die Struktur folgt der Organisation: Teamgröße, Zahl der Cluster und die Frage, wer welchen Teil verantwortet. Drei verbreitete Layouts:

- Monorepo mit einem Verzeichnis pro Cluster. Einfach und übersichtlich für ein Team und wenige Cluster.
- Geteilte Basis mit Cluster-Overlays. Sinnvoll, wenn mehrere ähnliche Cluster fast denselben Stack fahren und sich nur in Variablen unterscheiden.
- Ein Repository pro Team oder Tenant. Passt zu großen Organisationen mit getrennter Verantwortung und eigenen Review-Regeln.

Das hier referenzierte Repo nutzt die geteilte Basis. App-Definitionen liegen einmal unter `apps/`, die Cluster unter `clusters/` referenzieren sie und tragen nur ihre Auswahl und Variablen.

```text
kubernetes/
├── apps/          # HelmRelease und Manifeste je App
├── clusters/
│   ├── _base/     # gemeinsame Flux-Kustomizations
│   └── home-01/   # pro Cluster: Auswahl und Variablen
├── components/    # wiederverwendbare Kustomize-Components
├── crds/          # CRDs, getrennt ausgerollt
└── repositories/  # HelmRepository und OCIRepository
```

## Bootstrap: zwei Wege

Antipattern: Den Cluster-Zustand imperativ per `kubectl apply` oder über die Oberfläche herstellen. Der Zustand ist dann nirgends nachvollziehbar abgelegt.

Besser: Flux deklarativ bootstrappen, sodass Flux sich selbst aus Git verwaltet. Dafür gibt es zwei etablierte Wege.

Der erste ist die Flux-CLI. `flux bootstrap` committet die Flux-Komponenten samt einer `GitRepository` und einer Root-Kustomization ins Repo. Flux verwaltet danach auch seine eigene Installation.

```bash
flux bootstrap github \
  --owner=example \
  --repository=homelab-k8s \
  --branch=main \
  --path=kubernetes/clusters/home-01
```

Der zweite ist der Flux Operator. Eine `FluxInstance` beschreibt die Distribution, die aktiven Controller und die Sync-Quelle als Custom Resource. Der Lebenszyklus von Flux wird damit selbst deklarativ; Upgrades und mehrere Instanzen sind einfacher.

Beispiel aus dem Repo: Die `FluxInstance` setzt die Sync-Ref über eine Variable statt sie fest zu verdrahten.

```yaml
sync:
  kind: GitRepository
  url: https://github.com/example/homelab-k8s.git
  ref: ${FLUX_SYNC_REF}   # z. B. refs/heads/main oder refs/tags/v1.4.0
  path: kubernetes/clusters/home-01/flux
```

Die CLI ist der direkteste offizielle Einstieg. Der Operator lohnt sich, sobald der Flux-Lebenszyklus selbst versioniert und pro Cluster unterschiedlich gepinnt werden soll.

## Antipattern: Klartext-Secrets im Repo

Klartext-Secrets im Git sind das häufigste Antipattern. Jeder mit Repo-Zugriff liest sie, und sie bleiben für immer in der History.

Besser: Secrets verschlüsselt ablegen, etwa mit SOPS und age, mit Sealed Secrets oder über External Secrets aus einem Vault. SOPS mit age verschlüsselt nur die Wertfelder, sodass die Struktur im Diff lesbar bleibt.

```yaml
# .sops.yaml
creation_rules:
  - path_regex: '.*\.sops\.ya?ml$'
    encrypted_regex: '^(data|stringData)$'
    age:
      - "<age-recipient>"
```

Flux entschlüsselt zur Laufzeit mit dem privaten Schlüssel, der als Secret im Cluster liegt und nicht ins Repo gehört. Eine Kustomize-Component aktiviert die Entschlüsselung pro Kustomization, statt sie überall einzeln zu setzen.

## Antipattern: CRDs und App im selben Release

Wird eine CRD vom selben HelmRelease verwaltet, das auch die zugehörige Custom Resource anlegt, entsteht eine Race Condition: Die Resource wird angewendet, bevor ihre CRD registriert ist.

Besser: CRDs getrennt ausrollen. Das HelmRelease trägt `crds: Skip`, die CRDs laufen in einer eigenen Kustomization mit `wait: true`, und die App hängt per `dependsOn` daran.

```yaml
spec:
  dependsOn:
    - name: cert-manager-crds
  wait: true
```

Damit liegt der CRD-Lebenszyklus bei Flux und die Reihenfolge ist garantiert.

## Antipattern: Konfiguration je Cluster kopieren

Mehrere Cluster per Copy-and-paste zu pflegen führt zu Drift zwischen ihnen. Eine Änderung muss an mehreren Stellen nachgezogen werden und wird irgendwo vergessen.

Besser: eine geteilte Basis plus Variablen. `postBuild.substitute` setzt einzelne Werte, `substituteFrom` zieht sie aus einer ConfigMap oder einem Secret.

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

Beispiel aus dem Repo: Die clusterweiten Variablen stehen einmal in der Root-Kustomization und werden vererbt. Wiederkehrende Felder wie `sourceRef`, `interval`, `prune` und `wait` setzt eine Kustomize-Component zentral, sodass jede Kustomization sie erbt statt sie zu wiederholen.

## Antipattern: Versionen von Hand pflegen

Image-Tags und Chart-Versionen von Hand zu pflegen ist fehleranfällig; `:latest` macht den Zustand unreproduzierbar.

Besser: Updates automatisieren. Zwei Ansätze, je nach gewünschtem Fluss.

Flux kann Image-Tags selbst aktualisieren. Eine `ImageRepository` scannt die Registry, eine `ImagePolicy` wählt die Version, eine `ImageUpdateAutomation` committet zurück ins Repo. Ein Marker-Kommentar verknüpft das Feld mit der Policy.

```yaml
image: ghcr.io/example/app:1.4.0 # {"$imagepolicy": "flux-system:app-policy"}
```

Alternativ öffnet ein externes Tool wie Renovate einen Pull Request, ein Mensch prüft und merged. Beispiel aus dem Repo: Chart-Versionen tragen eine Renovate-Annotation und werden über Review aktualisiert.

```yaml
# renovate: datasource=helm depName=cert-manager registryUrl=https://charts.jetstack.io
CERT_MANAGER_VERSION: "v1.20.2"
```

Die Flux-Image-Automation passt, wenn neue App-Images automatisch einfließen sollen. Der PR-Weg passt, wenn jede Änderung über ein Review laufen soll. Beides ist GitOps-konform, weil die Änderung am Ende ein Commit ist.

## Zusammenfassung

- Flux gleicht das Cluster gegen Git ab; `prune` entfernt Drift, `wait` ermöglicht `dependsOn`
- Es gibt kein richtiges Layout; Monorepo, geteilte Basis oder Repo pro Team folgen den Bedürfnissen der Organisation
- Bootstrap deklarativ, entweder per `flux bootstrap` oder über den Flux Operator mit einer `FluxInstance`
- Die Sync-Ref pro Cluster über eine Variable wie `FLUX_SYNC_REF` pinnen
- Secrets verschlüsselt ablegen, etwa mit SOPS und age, nie im Klartext
- CRDs getrennt mit `wait: true` ausrollen, Apps per `dependsOn` anhängen
- Cluster-Duplikation über geteilte Basis plus `postBuild`-Variablen und Components vermeiden
- Versionen automatisiert aktualisieren, per Flux-Image-Automation oder per Renovate-PR
