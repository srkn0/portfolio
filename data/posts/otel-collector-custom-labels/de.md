---
title: "OpenTelemetry Collector: eigene Labels auf allen Signalen setzen"
description: "Einen statischen Identifier auf Logs, Metrics und Traces setzen, resource-Processor, Umgebungsvariablen in Kubernetes, attributes-Processor, transform-Processor mit OTTL."
tags: [opentelemetry, observability, kubernetes, collector]
date: 2025-05-20
---

## Ausgangslage

Ein statischer Identifier wie `project_id` oder `collector_id` soll auf allen Signalen liegen: Logs, Metrics und Traces. Der Contrib-Collector bietet dafür drei Wege.

## resource-Processor

Der `resource`-Processor fügt Resource-Attribute hinzu, die an alle Signale weitergereicht werden.

```yaml
processors:
  resource/add_project:
    attributes:
      - action: upsert
        key: project_id
        value: "<project-id>"

service:
  pipelines:
    traces:
      processors: [resource/add_project, memory_limiter, batch]
    metrics:
      processors: [resource/add_project, memory_limiter, batch]
    logs:
      processors: [resource/add_project, memory_limiter, batch]
```

`action: upsert` legt das Attribut an oder überschreibt einen vorhandenen Wert. Der Ansatz ist einfach und wirkt global auf alle Signale.

### Wert per Umgebungsvariable in Kubernetes

Der Wert kann aus einer Umgebungsvariable stammen. Das entkoppelt die Konfiguration vom konkreten Identifier.

```yaml
processors:
  resource/add_project:
    attributes:
      - action: upsert
        key: project_id
        value: "${env:PROJECT_ID}"
```

Die Variable wird im Pod gesetzt, etwa über das Deployment des Collectors.

```yaml
env:
  - name: PROJECT_ID
    value: "<project-id>"
```

## attributes-Processor

Für feinkörnigere Aktionen je Pipeline eignet sich der `attributes`-Processor. Er arbeitet auf Signal-Attributen statt auf Resource-Attributen.

```yaml
processors:
  attributes/add_project:
    actions:
      - action: upsert
        key: project_id
        value: "<project-id>"

service:
  pipelines:
    traces:
      processors: [attributes/add_project, memory_limiter, batch]
    metrics:
      processors: [attributes/add_project, memory_limiter, batch]
    logs:
      processors: [attributes/add_project, memory_limiter, batch]
```

Der Geltungsbereich wird über die Zuordnung zur jeweiligen Pipeline gesteuert, nicht über ein Signaltyp-Feld in der Processor-Definition. Neben `upsert` stehen `insert`, `update`, `delete` und `hash` zur Verfügung.

## transform-Processor mit OTTL

Bei bedingten oder berechneten Werten kommt der `transform`-Processor mit OTTL zum Einsatz. Die Statements werden je Signaltyp und Kontext angegeben.

```yaml
processors:
  transform/add_project:
    error_mode: ignore
    trace_statements:
      - context: span
        statements:
          - set(attributes["project_id"], "<project-id>")
    metric_statements:
      - context: datapoint
        statements:
          - set(attributes["project_id"], "<project-id>")
    log_statements:
      - context: log
        statements:
          - set(attributes["project_id"], "<project-id>")

service:
  pipelines:
    traces:
      processors: [transform/add_project, memory_limiter, batch]
    metrics:
      processors: [transform/add_project, memory_limiter, batch]
    logs:
      processors: [transform/add_project, memory_limiter, batch]
```

Diese Form ist ausführlicher, erlaubt aber Bedingungen und Wertetransformationen. Ein `where`-Ausdruck kann das `set` an eine Bedingung knüpfen.

## Zusammenfassung

- Der resource-Processor ist der erste Weg für globale, statische Resource-Attribute.
- `action: upsert` legt das Attribut an oder überschreibt es.
- Der Wert kann in Kubernetes per `${env:…}` aus einer Umgebungsvariable kommen.
- Der attributes-Processor bietet feinkörnige Aktionen je Pipeline (insert, upsert, update, delete).
- Der Geltungsbereich des attributes-Processors ergibt sich aus der Pipeline-Zuordnung.
- Der transform-Processor mit OTTL bleibt bedingten oder berechneten Werten vorbehalten.
