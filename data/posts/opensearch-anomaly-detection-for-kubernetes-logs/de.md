---
title: "Anomaly Detection für Kubernetes-Logs mit OpenSearch"
description: "Anomalieerkennung in Logs mit OpenSearch und dem Random Cut Forest Algorithmus, Setup mit Data Prepper und OpenTelemetry, das Anlegen eines Detectors, numerische Features und count() auf Kubernetes-Logs."
tags: [opensearch, observability, kubernetes, anomaly-detection]
date: 2025-08-18
---

## Anomalieerkennung in Logs

Anomalieerkennung in Logs gehört in den Bereich Machine Learning. Algorithmen lernen aus vorhandenen Daten typische Muster.

Auf Basis vorab definierter Kriterien werden anschließend ungewöhnliche Abweichungen erkannt.

## OpenSearch

OpenSearch ist eine quelloffene Search- und Observability-Suite und ein Fork von Elasticsearch und Kibana. Das Projekt wurde 2021 gestartet und wird unter anderem von AWS getragen.

OpenSearch enthält ein Anomaly-Detection-Feature, das kostenlos nutzbar ist. Der ELK-Stack bietet vergleichbare Machine-Learning-Funktionen nur mit Subscription.

Das Feature nutzt den Random Cut Forest (RCF) Algorithmus. Ausgewertet werden numerische Zeitreihen, aus denen der Algorithmus einen Anomaly Score berechnet. Der Score zeigt an, wie stark ein Datenpunkt vom gelernten Muster abweicht.

## Architektur

Logs, Metriken und Traces lassen sich aus einem OpenTelemetry-Setup über eine Zwischenkomponente, den Data Prepper, nach OpenSearch pushen.

Data Prepper ist eine Collector-Komponente von OpenSearch. Sie nimmt Logs, Metriken und Traces entgegen und erlaubt deren Transformation in einer Custom Pipeline, ähnlich dem OpenTelemetry Collector. Anschließend leitet sie die Daten an OpenSearch weiter.

![Architektur: OpenTelemetry Collector, Data Prepper und OpenSearch Cluster mit OpenSearch Dashboards als Frontend](/public/img/posts/opensearch-anomaly-detection/architecture.png)

Der OpenTelemetry Collector exportiert die Daten an die `otel_logs_source` des Data Prepper auf Port `21892`. Data Prepper leitet sie an ein OpenSearch-Cluster weiter. OpenSearch Dashboards dient als Frontend.

## Setup

Alle Komponenten lassen sich über Helm Charts deployen.

```bash
helm install opensearch opensearch-project/opensearch
helm install data-prepper opensearch-project/data-prepper
helm install opensearch-dashboards opensearch-project/opensearch-dashboards
```

Die Default-Helm-Values reichen für einen ersten Aufbau weitgehend aus. Die Data Prepper Pipeline aus der Architektur wird direkt im Helm Chart definiert.

### Initiales Admin-Passwort

Beim OpenSearch-Helm-Chart muss unter `extraEnvs` ein initiales Admin-Passwort gesetzt werden.

```yaml
extraEnvs:
  - name: OPENSEARCH_INITIAL_ADMIN_PASSWORD
    value: <password>
```

Ohne diese Variable startet das Cluster nicht. Der Wert ist ein Platzhalter und wird über ein Secret eingebunden.

## Logs einsehen und Setup prüfen

Die Data Prepper Pipeline schreibt alle Logs in einen Index, im Beispiel `all_logs`. Als Quelle dienen Kubernetes-Logs und -Events aus dem OpenTelemetry-Setup.

Unter Index Management zeigen die Mappings die indexierten Felder. Unter OpenSearch Dashboards > Discover lassen sich die Logs einsehen.

## Detector anlegen

Das Feature liegt unter OpenSearch Plugins > Anomaly Detection. Die Oberfläche beschreibt das Vorgehen zum Anlegen eines Detectors.

![Get-started-Ansicht des Anomaly-Detection-Plugins in OpenSearch Dashboards](/public/img/posts/opensearch-anomaly-detection/get-started.png)

### Schritt 1: Detector definieren

Im ersten Schritt wird der Detector definiert.

- Name des Detectors
- Source Index, in dem Anomalien gesucht werden, im Beispiel `all_logs`
- Timestamp-Feld zum Filtern, im Beispiel `time`

![Detail-Ansicht zur Definition eines Detectors mit Name, Source Index und Timestamp-Feld](/public/img/posts/opensearch-anomaly-detection/detector-details.png)

### Schritt 2: Modell konfigurieren

Im zweiten Schritt werden Features definiert. Ein Feature ist das Feld im Index, das auf Anomalien geprüft wird. Pro Detector sind bis zu fünf Features möglich.

Anomalieerkennung erfolgt immer auf numerischen Features. Zur Auswahl stehen die Aggregationsmethoden `average()`, `count()`, `sum()`, `min()` und `max()`.

Ablauf je Feld:

- Festlegen, für welche Felder Anomalien erfasst werden sollen.
- Prüfen, ob das Feld vom Typ number oder string ist.

Bei `Type: number` lässt sich das Feature direkt definieren und eine passende Aggregationsmethode wählen.

Bei `Type: string` ist nur `count()` ohne Vorarbeit nutzbar. Diese Methode zählt, wie oft ein Feld in Logs auftaucht, und liefert so ein numerisches Feature. Reichere numerische Features erfordern eine Ableitung von Metriken aus Logs in einer Data Prepper Pipeline (logs-to-metrics).

![Feature mit count() über das Kubernetes-Feld für den Container-Namen](/public/img/posts/opensearch-anomaly-detection/feature-count.png)

### Categorical Fields

Nach den Features lassen sich Categorical Fields setzen. Im Beispiel der Kubernetes-Logs dient der Kubernetes-Namespace als Categorical Field. Anomalien werden dadurch nach Namespace gruppiert.

### Schritt 3: Detector-Jobs einrichten

Die Vorauswahl wird beibehalten: Start real-time detector automatically.

### Schritt 4: Prüfen und anlegen

Im letzten Schritt wird der Detector geprüft und über Create Detector angelegt.

## Beispiel Detector

Der Beispiel-Detector arbeitet mit vier Features, jeweils mit der Aggregationsmethode `count()`.

```text
Container-loggt-ungewoehnlich-oft   resource.attributes.k8s@container@name.keyword
Deployment-loggt-ungewoehnlich-oft  resource.attributes.k8s@deployment@name.keyword
Statefulset-loggt-ungewoehnlich-oft resource.attributes.k8s@statefulset@name.keyword
Ungewoehnlich-viele-Events          log.attributes.event@domain.keyword
```

Die kategorische Gruppierung erfolgt auf Namespace-Basis. Im Dashboard unter Anomaly Overview lassen sich einzelne Kacheln öffnen, der Feature Breakdown zeigt die Anomalien je Feature.

## Numerische Features bei Logs

Anomaly Detection in OpenSearch erfordert numerische Features. Bei Logs sind diese nicht immer vorhanden.

Kubernetes-Logs aus dem OpenTelemetry Collector tragen keine numerischen Felder. Felder wie `droppedAttributesCount`, `flags` und `severityNumber` werden zwar indexiert, bleiben für Kubernetes-Logs und -Events aber leer. Nutzbar ist daher nur `count()`, also das Log-Volumen pro Feld.

Anwendungslogs können `severityNumber` als echtes numerisches Feature enthalten, das als Grundlage für einen Detector dienen kann.

Weitere numerische Features erfordern Data Prepper Pipelines, die Logs zu Metriken konvertieren. Dabei wird festgelegt, welche Felder relevant sind und wie daraus aussagekräftige Zahlenwerte entstehen.

Bei Metriken liegen alle Werte bereits im Zahlenformat vor, was den Aufbau eines Detectors vereinfacht.

## Zusammenfassung

- OpenSearch ist ein Fork von Elasticsearch und Kibana und bringt Anomaly Detection als kostenloses Feature mit.
- Das Feature nutzt den Random Cut Forest Algorithmus und berechnet einen Anomaly Score je Datenpunkt.
- Logs, Metriken und Traces gelangen über Data Prepper aus einem OpenTelemetry-Setup nach OpenSearch.
- Das OpenSearch-Helm-Chart erfordert `OPENSEARCH_INITIAL_ADMIN_PASSWORD` unter `extraEnvs`.
- Ein Detector benötigt Source Index, Timestamp-Feld und numerische Features mit einer Aggregationsmethode.
- Kubernetes-Logs aus dem OpenTelemetry Collector tragen keine numerischen Felder, nutzbar ist nur `count()`.
- Anwendungslogs können `severityNumber` als numerisches Feature bereitstellen.
- Reichere numerische Features entstehen durch logs-to-metrics in einer Data Prepper Pipeline.
- Categorical Fields wie der Namespace gruppieren die erkannten Anomalien.
