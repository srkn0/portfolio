---
title: "OpenTelemetry Collector: setting custom labels on all signals"
description: "Setting a static identifier on logs, metrics and traces, resource processor, environment variables in Kubernetes, attributes processor, transform processor with OTTL."
tags: [opentelemetry, observability, kubernetes, collector]
date: 2025-05-20
---

## Starting point

A static identifier such as `project_id` or `collector_id` should be present on all signals: logs, metrics and traces. The Contrib Collector offers three ways to do this.

## resource processor

The `resource` processor adds resource attributes that are propagated to all signals.

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

`action: upsert` creates the attribute or overwrites an existing value. The approach is simple and applies globally to all signals.

### Value from an environment variable in Kubernetes

The value can come from an environment variable. This decouples the configuration from the concrete identifier.

```yaml
processors:
  resource/add_project:
    attributes:
      - action: upsert
        key: project_id
        value: "${env:PROJECT_ID}"
```

The variable is set in the pod, for example through the collector deployment.

```yaml
env:
  - name: PROJECT_ID
    value: "<project-id>"
```

## attributes processor

For fine-grained actions per pipeline, the `attributes` processor is suitable. It operates on signal attributes rather than resource attributes.

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

The scope is controlled by assigning the processor to the respective pipeline, not by a signal-type field in the processor definition. Besides `upsert`, the actions `insert`, `update`, `delete` and `hash` are available.

## transform processor with OTTL

For conditional or computed values, the `transform` processor with OTTL is used. The statements are given per signal type and context.

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

This form is more verbose but allows conditions and value transformations. A `where` clause can tie the `set` to a condition.

## Summary

- The resource processor is the first choice for global, static resource attributes.
- `action: upsert` creates the attribute or overwrites it.
- In Kubernetes the value can come from an environment variable via `${env:…}`.
- The attributes processor offers fine-grained actions per pipeline (insert, upsert, update, delete).
- The scope of the attributes processor follows from the pipeline assignment.
- The transform processor with OTTL is reserved for conditional or computed values.
