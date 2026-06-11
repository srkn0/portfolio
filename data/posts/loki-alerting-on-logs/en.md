---
title: "Alerting on logs with the Loki ruler"
description: "The Loki ruler, rulerConfig in the Helm chart values, a Kustomize JSON6902 patch to mount a rules ConfigMap into the Loki StatefulSet, a ConfigMap with LogQL alerting rules and the per-tenant rules path layout."
tags: [loki, observability, alerting, kubernetes]
date: 2025-06-26
---

## The Loki ruler

The Loki ruler evaluates Prometheus-style alerting rules whose expression is a LogQL query. Matches are sent to Alertmanager as alerts.

This makes logs a direct alert source, without converting them to metrics first. Evaluation runs in the Loki ruler component.

## rulerConfig in the Helm chart values

In the Helm chart values a `rulerConfig` block enables the ruler.

```yaml
rulerConfig:
  alertmanager_url: http://alertmanager-operated.example.com:9093
  ring:
    kvstore:
      store: inmemory
  enable_api: true
```

`alertmanager_url` points to the Alertmanager that receives the alerts. `ring.kvstore.store: inmemory` keeps the ruler ring in memory and is sufficient for a single replica. `enable_api: true` exposes the ruler API.

## Mounting rules via a Kustomize patch

The rules live in a ConfigMap and are mounted into the Loki StatefulSet through a Kustomize JSON6902 patch.

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

The first patch appends a `volumeMount` to the Loki container, the second a `volume` that references the ConfigMap. `subPath: rules.yaml` mounts only that one key as a file.

### Per-tenant rules path layout

Loki expects rules under `<rules-path>/<tenant>/`. The tenant part of the path is required.

```text
/var/loki/rules/<tenant>/rules.yaml
```

In single-tenant mode `fake` is the default tenant id. That is why the file lands under `fake/`, that is `/var/loki/rules/fake/rules.yaml`.

## ConfigMap with the alerting rules

The ConfigMap holds the `rules.yaml` with a `groups`/`rules` structure.

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
              summary: Elevated rate of error log lines in example-app
```

The `expr` expression is a LogQL query. The inner `rate(...)` expression measures the rate of matching log lines over five minutes, `sum(...)` aggregates them, and the threshold `> 0.5` fires the alert. `for: 5m` requires the condition to hold for five minutes before the alert moves from pending to firing. `labels` and `annotations` are attached to the alert.

## Summary

- The Loki ruler evaluates alerting rules with LogQL expressions and sends alerts to Alertmanager.
- Alerting on logs happens without converting them to metrics first.
- `rulerConfig` enables the ruler with `alertmanager_url`, `ring.kvstore.store: inmemory` and `enable_api: true`.
- A Kustomize JSON6902 patch mounts the rules ConfigMap as a `volumeMount` and `volume` into the Loki StatefulSet.
- Loki expects rules under `<rules-path>/<tenant>/`.
- `fake` is the default tenant id in single-tenant mode, hence the `fake/` path.
- A rule consists of `expr` (LogQL), `for`, `labels` and `annotations`.
