---
title: "Entwicklung gegen ein prod-nahes Cluster mit DevSpace und KinD"
description: "Lokale Entwicklung gegen ein echtes Kubernetes statt gegen Mocks: ein KinD-Cluster mit demselben Helm-Chart wie in Produktion, ein Bootstrap-Skript in einem Befehl und der DevSpace-Hot-Reload-Loop mit File-Sync über mehrere Dienste. Warum Prod-Parität die Entwicklung konsistenter macht."
tags: [kubernetes, devspace, kind, developer-experience, helm]
date: 2026-05-22
---

## Das Problem mit lokaler Entwicklung

Lokale Umgebungen weichen von Produktion ab. Ein Docker-Compose-Stack oder Mocks bilden Datenbank und HTTP nach, aber nicht das, worauf der Code in Produktion trifft: Kubernetes-Manifeste, CRDs, RBAC, Service-Discovery und Ingress-Routing. Fehler tauchen dann erst bei der Bereitstellung auf.

Bei einem Kubernetes-Operator ist die Lücke besonders groß. Ein Operator reconciled Custom Resources über die Kubernetes-API. Ohne echte API gibt es nichts zu reconcilen. Der Operator lässt sich nur sinnvoll gegen ein laufendes Cluster entwickeln.

Das hier beschriebene Setup stammt aus dem Projekt [kubeplate](/projects/kubeplate) und löst das, indem lokal dasselbe läuft wie in Produktion.

## Ein echtes Cluster lokal mit KinD

KinD startet ein vollständiges Kubernetes in Docker-Containern. In dieses Cluster wird dasselbe Helm-Chart deployt, das auch in Produktion läuft: dieselben CRDs, dieselben RBAC-Regeln, dieselbe Datenbank, dasselbe Routing.

Das ist Parität statt Simulation. Was lokal funktioniert, funktioniert auch im Cluster, weil es dasselbe Cluster ist.

```yaml
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080   # Traefik HTTP
        hostPort: 80
      - containerPort: 30443   # Traefik HTTPS
        hostPort: 443
```

Die Port-Mappings reichen Host-Port 80 und 443 an den Ingress im Cluster durch. Damit ist das pfadbasierte Routing lokal dasselbe wie in Produktion: `/api` geht an den Server, `/` an die Weboberfläche.

## Bootstrap in einem Befehl

Ein Skript legt das Cluster idempotent an. Es prüft, ob das Cluster existiert, erzeugt es bei Bedarf, installiert optional Gateway API und Traefik und rollt das Helm-Chart aus.

```bash
./dev/bootstrap-kind.sh
```

Idempotenz ist hier der Punkt. Das Skript lässt sich beliebig oft ausführen; ein bestehendes Cluster wird nicht neu gebaut. Onboarding ist damit ein Befehl, nicht eine Seite Anleitung.

## Der Hot-Reload-Loop mit DevSpace

`devspace dev` ersetzt die Anwendungs-Pods durch Entwicklungs-Container und stellt eine Verbindung zwischen Arbeitsverzeichnis und Cluster her. Im Fall von kubeplate laufen drei Dienste parallel: die Weboberfläche, der Server und der Operator.

```yaml
dev:
  web:
    sync:
      - path: ./:/app
    ports:
      - port: "3000:3000"
    command: ["sh", "-c", "pnpm install && cd apps/web && next dev"]
```

Drei Mechanismen greifen ineinander:

- File-Sync spiegelt geänderte Dateien sofort in den Pod, ohne ein neues Image zu bauen.
- Jeder Stack nutzt Hot-Reload: die Weboberfläche über den Dev-Server, der Server über einen Watch-Prozess und der Operator über `air`, das die Go-Binary bei jeder Änderung neu baut.
- Caches wie `node_modules` und der Go-Modul-Cache liegen auf persistenten Volumes und überstehen Pod-Neustarts.

Eine Codeänderung ist damit Sekunden später im Cluster aktiv. Der innere Loop ist so kurz wie bei rein lokaler Entwicklung, läuft aber gegen echtes Kubernetes.

## Gegen echte Infrastruktur entwickeln

Der Gewinn ist Konsistenz. Der Operator reconciled echte Custom Resources, RBAC und Service-Discovery verhalten sich wie in Produktion, und das Ingress-Routing ist dasselbe. Es gibt keine Klasse von Fehlern mehr, die erst nach dem Deploy auftaucht, weil lokal etwas fehlte.

Konsistenz gilt auch im Team. Jeder entwickelt gegen dieselbe Cluster-Definition und dasselbe Helm-Chart. Die gepinnte Toolchain und das Bootstrap-Skript sorgen dafür, dass die Umgebung auf jeder Maschine gleich aussieht. „Bei mir läuft es" verliert seine Bedeutung, wenn „bei mir" überall dasselbe ist.

## Workflow

```bash
pnpm install
./dev/bootstrap-kind.sh
devspace dev
```

Drei Befehle vom Klon bis zum Live-Coding: Abhängigkeiten installieren, das Cluster hochziehen, den Hot-Reload-Loop starten. Danach werden Änderungen an Weboberfläche, Server und Operator sofort im Cluster wirksam.

## Zusammenfassung

- Lokale Mocks und Compose-Stacks weichen von Produktion ab; bei Operatoren fehlt ohne echte API die Grundlage
- KinD startet ein echtes Kubernetes lokal, in das dasselbe Helm-Chart wie in Produktion deployt wird
- Port-Mappings bringen echtes Ingress-Routing auf den lokalen Host
- Ein idempotentes Bootstrap-Skript macht das Onboarding zu einem Befehl
- `devspace dev` synchronisiert Dateien live und nutzt für jeden Dienst Hot-Reload, ohne Image-Rebuild
- Persistente Caches halten den inneren Loop kurz
- Entwicklung gegen echte CRDs, RBAC und Routing verschiebt keine Fehlerklasse mehr auf den Deploy
- Eine geteilte Cluster-Definition macht die Umgebung im Team konsistent
