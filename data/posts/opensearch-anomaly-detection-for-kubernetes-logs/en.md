---
title: "Anomaly detection for Kubernetes logs with OpenSearch"
description: "Anomaly detection in logs with OpenSearch and the Random Cut Forest algorithm, setup with Data Prepper and OpenTelemetry, creating a detector, numeric features and count() on Kubernetes logs."
tags: [opensearch, observability, kubernetes, anomaly-detection]
date: 2025-08-18
---

## Anomaly detection in logs

Anomaly detection in logs belongs to the field of machine learning. Algorithms learn typical patterns from existing data.

Based on criteria defined in advance, unusual deviations are then detected.

## OpenSearch

OpenSearch is an open-source search and observability suite and a fork of Elasticsearch and Kibana. The project was started in 2021 and is backed by AWS among others.

OpenSearch includes an anomaly-detection feature that is free to use. The ELK stack offers comparable machine-learning functions only with a subscription.

The feature uses the Random Cut Forest (RCF) algorithm. It evaluates numeric time series, from which the algorithm computes an anomaly score. The score indicates how far a data point deviates from the learned pattern.

## Architecture

Logs, metrics and traces can be pushed from an OpenTelemetry setup to OpenSearch through an intermediate component, Data Prepper.

Data Prepper is a collector component of OpenSearch. It receives logs, metrics and traces and allows their transformation in a custom pipeline, similar to the OpenTelemetry Collector. It then forwards the data to OpenSearch.

![Architecture: OpenTelemetry Collector, Data Prepper and OpenSearch cluster with OpenSearch Dashboards as the frontend](/public/img/posts/opensearch-anomaly-detection/architecture.png)

The OpenTelemetry Collector exports the data to the `otel_logs_source` of Data Prepper on port `21892`. Data Prepper pushes it to an OpenSearch cluster. OpenSearch Dashboards serves as the frontend.

## Setup

All components can be deployed via Helm charts.

```bash
helm install opensearch opensearch-project/opensearch
helm install data-prepper opensearch-project/data-prepper
helm install opensearch-dashboards opensearch-project/opensearch-dashboards
```

The default Helm values are largely sufficient for an initial setup. The Data Prepper pipeline from the architecture is defined directly in the Helm chart.

### Initial admin password

The OpenSearch Helm chart requires an initial admin password under `extraEnvs`.

```yaml
extraEnvs:
  - name: OPENSEARCH_INITIAL_ADMIN_PASSWORD
    value: <password>
```

Without this variable the cluster does not start. The value is a placeholder and is provided through a secret.

## Viewing logs and verifying the setup

The Data Prepper pipeline writes all logs into an index, in the example `all_logs`. Kubernetes logs and events from the OpenTelemetry setup serve as the source.

Under Index Management the mappings show the indexed fields. Under OpenSearch Dashboards > Discover the logs can be viewed.

## Creating a detector

The feature is located under OpenSearch Plugins > Anomaly Detection. The interface describes how to create a detector.

![Get-started view of the anomaly-detection plugin in OpenSearch Dashboards](/public/img/posts/opensearch-anomaly-detection/get-started.png)

### Step 1: Define detector

The first step defines the detector.

- Name of the detector
- Source index in which anomalies are searched, in the example `all_logs`
- Timestamp field for filtering, in the example `time`

![Detail view for defining a detector with name, source index and timestamp field](/public/img/posts/opensearch-anomaly-detection/detector-details.png)

### Step 2: Configure model

The second step defines features. A feature is the field in the index that is checked for anomalies. Up to five features are possible per detector.

Anomaly detection always operates on numeric features. The aggregation methods available are `average()`, `count()`, `sum()`, `min()` and `max()`.

Procedure per field:

- Decide for which fields anomalies should be captured.
- Check whether the field is of type number or string.

For `Type: number` the feature can be defined directly and a suitable aggregation method chosen.

For `Type: string` only `count()` is usable without preparation. This method counts how often a field appears in logs and thus yields a numeric feature. Richer numeric features require deriving metrics from logs in a Data Prepper pipeline (logs-to-metrics).

![Feature with count() over the Kubernetes field for the container name](/public/img/posts/opensearch-anomaly-detection/feature-count.png)

### Categorical fields

After the features, categorical fields can be set. In the example of the Kubernetes logs, the Kubernetes namespace serves as the categorical field. Anomalies are thereby grouped by namespace.

### Step 3: Set up detector jobs

The preselection is kept: start real-time detector automatically.

### Step 4: Review and create

In the final step the detector is reviewed and created via Create Detector.

## Example detector

The example detector works with four features, each with the aggregation method `count()`.

```text
Container-loggt-ungewoehnlich-oft   resource.attributes.k8s@container@name.keyword
Deployment-loggt-ungewoehnlich-oft  resource.attributes.k8s@deployment@name.keyword
Statefulset-loggt-ungewoehnlich-oft resource.attributes.k8s@statefulset@name.keyword
Ungewoehnlich-viele-Events          log.attributes.event@domain.keyword
```

The categorical grouping is on a namespace basis. In the dashboard under Anomaly Overview individual tiles can be opened, and the feature breakdown shows the anomalies per feature.

## Numeric features for logs

Anomaly detection in OpenSearch requires numeric features. For logs these are not always present.

Kubernetes logs from the OpenTelemetry Collector carry no numeric fields. Fields such as `droppedAttributesCount`, `flags` and `severityNumber` are indexed but remain empty for Kubernetes logs and events. Only `count()` is usable, that is the log volume per field.

Application logs may contain `severityNumber` as a real numeric feature, on which a detector can be built.

Further numeric features require Data Prepper pipelines that convert logs to metrics. This defines which fields are relevant and how meaningful numeric values are derived from them.

For metrics all values are already in numeric form, which simplifies building a detector.

## Summary

- OpenSearch is a fork of Elasticsearch and Kibana and ships anomaly detection as a free feature.
- The feature uses the Random Cut Forest algorithm and computes an anomaly score per data point.
- Logs, metrics and traces reach OpenSearch from an OpenTelemetry setup via Data Prepper.
- The OpenSearch Helm chart requires `OPENSEARCH_INITIAL_ADMIN_PASSWORD` under `extraEnvs`.
- A detector needs a source index, a timestamp field and numeric features with an aggregation method.
- Kubernetes logs from the OpenTelemetry Collector carry no numeric fields, only `count()` is usable.
- Application logs can provide `severityNumber` as a numeric feature.
- Richer numeric features come from logs-to-metrics in a Data Prepper pipeline.
- Categorical fields such as the namespace group the detected anomalies.
