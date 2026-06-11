---
title: "Alerting auf Logs mit dem Loki-Ruler"
description: "Der Loki-Ruler, rulerConfig in den Helm-Chart-Values, ein Kustomize-JSON6902-Patch zum Mounten einer Rules-ConfigMap in das Loki-StatefulSet, eine ConfigMap mit LogQL-Alerting-Rules und das Rules-Pfad-Layout pro Tenant."
tags: [loki, observability, alerting, kubernetes]
date: 2025-06-26
---

## Der Loki-Ruler

Der Loki-Ruler wertet Prometheus-artige Alerting-Rules aus, deren Ausdruck eine LogQL-Query ist. Treffer werden als Alerts an Alertmanager geschickt.

Damit lassen sich Logs direkt als Alert-Quelle nutzen, ohne sie zuvor in Metriken umzuwandeln. Die Auswertung läuft im Ruler-Komponente von Loki.

## rulerConfig in den Helm-Chart-Values

In den Helm-Chart-Values aktiviert ein `rulerConfig`-Block den Ruler.

```yaml
rulerConfig:
  alertmanager_url: http://alertmanager-operated.example.com:9093
  ring:
    kvstore:
      store: inmemory
  enable_api: true
```

`alertmanager_url` verweist auf den Alertmanager, der die Alerts empfängt. `ring.kvstore.store: inmemory` hält den Ruler-Ring im Speicher und genügt für eine einzelne Replica. `enable_api: true` schaltet die Ruler-API frei.

## Rules per Kustomize-Patch mounten

Die Rules liegen in einer ConfigMap und werden über einen Kustomize-JSON6902-Patch in das Loki-StatefulSet gemountet.

```yaml
patches:
  - target:
      kind: StatefulSet
      name: loki
      namespace: loki
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/volumeMounts/-
        value:
          name: loki-ruler-cm
          mountPath: /var/loki/rules/fake/rules.yaml
          subPath: rules.yaml
      - op: add
        path: /spec/template/spec/volumes/-
        value:
          name: loki-ruler-cm
          configMap:
            name: loki-ruler-cm
```

Der erste Patch hängt einen `volumeMount` an den Loki-Container an, der zweite ein `volume`, das die ConfigMap referenziert. `subPath: rules.yaml` mountet nur den einen Schlüssel als Datei.

### Rules-Pfad-Layout pro Tenant

Loki erwartet Rules unter `<rules-path>/<tenant>/`. Der Tenant-Teil des Pfades ist erforderlich.

```text
/var/loki/rules/<tenant>/rules.yaml
```

Im Single-Tenant-Modus ist `fake` die Standard-Tenant-ID. Deshalb liegt die Datei unter `fake/`, also `/var/loki/rules/fake/rules.yaml`.

## ConfigMap mit den Alerting-Rules

Die ConfigMap enthält die `rules.yaml` mit einer `groups`/`rules`-Struktur.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-ruler-cm
  namespace: loki
data:
  rules.yaml: |
    groups:
      - name: log-alerts
        rules:
          - alert: HighErrorLogRate
            expr: |
              sum(rate({k8s_namespace_name="example-app"} |= `error` [5m])) > 0.5
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: Erhöhte Rate an Fehler-Logzeilen in example-app
```

Der `expr`-Ausdruck ist eine LogQL-Query. Der innere `rate(...)`-Ausdruck misst die Rate passender Logzeilen über fünf Minuten, `sum(...)` aggregiert sie, der Schwellwert `> 0.5` löst den Alert aus. `for: 5m` verlangt, dass die Bedingung fünf Minuten anhält, bevor der Alert von pending auf firing wechselt. `labels` und `annotations` werden an den Alert angehängt.

## Zusammenfassung

- Der Loki-Ruler wertet Alerting-Rules mit LogQL-Ausdrücken aus und sendet Alerts an Alertmanager.
- Alerting auf Logs erfolgt ohne vorherige Umwandlung in Metriken.
- `rulerConfig` aktiviert den Ruler mit `alertmanager_url`, `ring.kvstore.store: inmemory` und `enable_api: true`.
- Ein Kustomize-JSON6902-Patch mountet die Rules-ConfigMap als `volumeMount` und `volume` in das Loki-StatefulSet.
- Loki erwartet Rules unter `<rules-path>/<tenant>/`.
- `fake` ist die Standard-Tenant-ID im Single-Tenant-Modus, daher der `fake/`-Pfad.
- Eine Rule besteht aus `expr` (LogQL), `for`, `labels` und `annotations`.
