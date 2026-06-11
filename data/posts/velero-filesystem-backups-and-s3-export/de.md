---
title: "Velero: Filesystem-Backups und S3-Export"
description: "Aufbau eines Velero-Backups, BackupStorageLocation und VolumeSnapshotLocation, Filesystem-Backups, Export von Volume-Daten nach S3 mit snapshot-move-data, Immutability, Verschlüsselung, RestoreResourceModifiers, Migrationen, Helm-Setup und Monitoring."
tags: [velero, kubernetes, backup, s3]
date: 2025-09-15
---

## Aufbau

Velero besteht aus einem Server/Controller im Cluster und einer CLI. Aufgabe ist Backup und Restore von Kubernetes-Ressourcen und Volumes.

```bash
velero backup create my-backup
```

Velero arbeitet auf Basis von CRDs. Backups, Restores und Schedules sind also Custom Resources im Cluster. Backups laufen on-demand oder per Schedule und tragen eine optionale TTL für Expiration und Retention.

Ein Velero-Backup besteht aus zwei Teilen:

- Metadaten und Manifeste in einem Object Store (Single Source of Truth)
- Volume-Daten, entweder über einen Cloud-Provider mit VolumeSnapshot-Support oder über ein Filesystem-Backup

## Ablauf eines Backups

Der Ablauf ist in der Velero-Dokumentation unter `how-velero-works` beschrieben.

```bash
velero backup create my-backup --include-namespaces app
```

Ablauf: Die CLI erzeugt eine Backup-CR im aktuellen Kubernetes-Kontext. Der BackupController erkennt die CR, validiert sie und holt die Manifeste von der Kubernetes-API. Die Manifeste werden in die `BackupStorageLocation` geladen. Für PersistentVolumes wählt der Controller je nach Konfiguration die passende Location und legt dort die Volume-Daten ab. Volume-Snapshots lassen sich mit `--snapshot-volumes=false` deaktivieren.

## Backup- und Volume-Locations

Da ein Backup in zwei Teile zerfällt, gibt es zwei Custom Resources für die Locations.

```yaml
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: default
  namespace: velero
spec:
  provider: aws
  objectStorage:
    bucket: my-velero-backups
  config:
    region: <region>
```

`BackupStorageLocation`: ein S3-Bucket, in dem alle Velero-Daten wie YAML und Metadaten liegen. Single Source of Truth.

`VolumeSnapshotLocation`: der Ort, an dem die Volume-Snapshots liegen. Die Metadaten im `BackupStorageLocation` referenzieren auf diesen Ort. Diese Location ist nur für Cloud-Provider gedacht. In einem Cloud-Cluster stellt der Provider eigene SnapshotClasses bereit, und ein providerspezifisches Plugin macht diese in Velero nutzbar.

Für Filesystem-Backups wird keine `VolumeSnapshotLocation` benötigt. Dort nutzt Velero die Kubernetes-internen VolumeSnapshotClasses.

## Filesystem-Backups

Ein Filesystem-Backup benötigt nur eine `BackupStorageLocation` und keine `VolumeSnapshotLocation`. Die Metadaten liegen im S3-Object-Store, die Volume-Daten werden über das Dateisystem gesichert, etwa über Ceph.

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: cephfs-snapclass
  labels:
    velero.io/csi-volumesnapshot-class: "true"
driver: cephfs.csi.ceph.com
deletionPolicy: Retain
```

Velero nutzt dafür eine `VolumeSnapshotClass` des external-snapshotter (`snapshot.storage.k8s.io`). Das Label `velero.io/csi-volumesnapshot-class: "true"` markiert die für Velero relevante Klasse.

Konvention: Ist keine `VolumeSnapshotLocation` definiert und sind Filesystem-Backups konfiguriert, nutzt Velero standardmäßig die Filesystem-Backups, da nichts anderes verfügbar ist. Die folgenden Steueroptionen sind in diesem Fall irrelevant.

### Opt-in

Sind sowohl `VolumeSnapshotLocation` als auch Filesystem-Backups aktiviert, steuert eine Annotation die Auswahl pro Volume.

```yaml
metadata:
  annotations:
    backup.velero.io/backup-volumes: pvc-volume,emptydir-volume
```

Standardmäßig macht Velero ein Filesystem-Backup nur, wenn das Kubernetes-Objekt diese Annotation trägt. Volumes ohne Annotation werden für normale VolumeSnapshots in Betracht gezogen, wenn `--snapshot-volumes` nicht auf `false` steht und das Volume snapshot-fähig ist.

Daraus folgt:

- Volumes mit und ohne Annotation, `--snapshot-volumes=false`: nur Filesystem-Backups.
- Volumes mit und ohne Annotation, `--snapshot-volumes=true` (default): Filesystem-Backup für annotierte Volumes, VolumeSnapshots für die übrigen.

### Opt-out

```yaml
metadata:
  annotations:
    backup.velero.io/backup-volumes-excludes: log-volume
```

Mit `backup.velero.io/backup-volumes-excludes` werden einzelne Volumes vom Filesystem-Backup ausgeschlossen.

### Default auf Filesystem-Backup

```bash
velero install --default-volumes-to-fs-backup
```

Mit `--default-volumes-to-fs-backup=true` ist der Default immer Filesystem-Backup. Geht ein Backup an einen Cloud-Provider, muss die `VolumeSnapshotLocation` dann explizit gesetzt werden.

## Volume-Daten nach S3 exportieren

Ein S3-kompatibler Store ohne Velero-Volume-Snapshot-Plugin kann Volume-Daten nicht direkt aufnehmen. Velero backuppt Volumes entweder über eine `VolumeSnapshotLocation` oder über Filesystem-Backups.

Problem: Stellt ein Provider nur S3 bereit und hat kein eigenes Snapshot-Plugin für Kubernetes, gibt es kein passendes Velero-Provider-Plugin für die Volume-Snapshots.

Lösung: Volume-Daten aus einem Filesystem-Backup lassen sich nach dem Backup in einen Object Store verschieben.

```bash
velero backup create my-backup --snapshot-move-data
```

Das Backup wird mit `--snapshot-move-data` erstellt. Standardmäßig landen die Volume-Daten im gleichen Bucket wie die `BackupStorageLocation`. Ein benutzerdefinierter Data Mover legt das Ziel genauer fest.

```bash
velero backup create my-backup --snapshot-move-data --data-mover <data-mover-name>
```

Damit konfiguriert das AWS-Plugin den S3-Store als `BackupStorageLocation`, und über Filesystem-Backup plus `snapshot-move-data` werden die Volume-Daten nach S3 geschoben.

## OpenStack-Provider

Für Cloud-Cluster auf OpenStack lässt sich das `velero-plugin-for-openstack` nutzen, das den Cinder Volume Snapshotter ansteuert.

```yaml
initContainers:
  - name: velero-plugin-for-openstack
    image: velero/velero-plugin-for-openstack:v0.8.0
    volumeMounts:
      - mountPath: /target
        name: plugins
```

In diesem Fall dient das S3 des Providers als `BackupStorageLocation`, und Cinder liefert die `VolumeSnapshotLocation`.

## RestoreResourceModifiers

RestoreResourceModifiers manipulieren Kubernetes-Ressourcen vor dem Restore. Sie liegen in einer ConfigMap und werden beim Restore referenziert.

```bash
velero restore create --from-backup my-backup --resource-modifier-configmap my-modifiers
```

Beim Restore kann immer nur eine ConfigMap referenziert werden. Alle Transformationen müssen in dieser ConfigMap stehen. Ein Overlay-Prinzip, das mehrere Sets zusammensetzt, ist nicht möglich.

## Migrationen

Ein vollständiges Backup und Restore eines Clusters wird als Migration bezeichnet.

```bash
velero backup create cluster-full
velero restore create --from-backup cluster-full
```

Nativ funktioniert das nur, wenn sich beide Cluster auf derselben Cloud-Plattform bewegen. Eine Migration zwischen unterschiedlichen Plattformen erfordert Filesystem-Backups.

## Immutability

Ab Version 1.11 funktionieren Velero-Backups nicht zuverlässig, wenn der Ziel-Object-Store mit Immutability konfiguriert ist.

```text
Cannot support backup data immutability
```

Das betrifft die `BackupStorageLocation`. Sie ist die Single Source of Truth, in die Velero den Status der Backups laufend schreibt. Für diesen Bucket lässt sich keine Immutability einrichten.

Ein separater Bucket für die exportierten Volume-Daten kann hingegen immutable sein. Folge wäre lediglich, dass Velero ein Volume erst nach Ablauf der Immutability-Dauer löschen kann. Wird die TTL des Backups gleich der Immutability-Dauer des Buckets gesetzt, entsteht dieses Problem nicht.

## Verschlüsselung

Die Metadaten und YAML-Manifeste liegen in der `BackupStorageLocation` unverschlüsselt. Die Volume-Snapshots sind verschlüsselt.

```yaml
credentials:
  existingSecret: velero-repo-credentials
```

Der Schlüssel liegt in einem globalen Secret, das vom Helm-Chart deployed wird. Dieses backupRepository-Secret verschlüsselt alle backupRepositories.

## Helm-Setup

Velero lässt sich über ein Helm-Chart deployen.

```bash
helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts/
helm install velero vmware-tanzu/velero --namespace velero --values values.yaml
```

Notwendige Konfiguration für Filesystem-Backups in den Helm-Values:

```yaml
configuration:
  features: EnableCSI
deployNodeAgent: true
nodeAgent:
  tolerations:
    - key: "node-role.kubernetes.io/control-plane"
      operator: "Exists"
      effect: "NoSchedule"
```

`features: EnableCSI` aktiviert die CSI-Unterstützung, `deployNodeAgent: true` startet den Node-Agent für Filesystem-Backups. Die Toleration erlaubt Backups von Workloads auf der Control Plane. Die genutzten VolumeSnapshotClasses brauchen das Label `velero.io/csi-volumesnapshot-class: "true"`.

Die Provider-Plugins werden als Init-Container eingebunden:

```yaml
initContainers:
  - name: velero-plugin-for-aws
    image: velero/velero-plugin-for-aws:v1.12.2
    imagePullPolicy: IfNotPresent
    volumeMounts:
      - mountPath: /target
        name: plugins
```

Die `BackupStorageLocation` wird unter `configuration.backupStorageLocation` konfiguriert:

```yaml
configuration:
  backupStorageLocation:
    - name: default
      provider: aws
      bucket: my-velero-backups
      default: true
      accessMode: ReadWrite
      credential:
        name: bsl-creds
        key: cloud
      config:
        region: <region>
        s3Url: http://minio.minio.svc:9000
        s3ForcePathStyle: true
```

`s3ForcePathStyle: true` stellt den AWS-SDK-Adressstil auf path-style um, was S3-kompatible Stores wie MinIO erwarten. Für das OpenStack-Szenario wird zusätzlich eine `VolumeSnapshotLocation` unter `configuration.volumeSnapshotLocation` gesetzt.

## Monitoring

Velero exponiert Prometheus-Metriken auf Port 8085.

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

Der Endpoint wird per ServiceMonitor gescraped. Auf Basis der Metriken kann Alertmanager benachrichtigen. Velero pflegt ein eigenes Grafana-Dashboard für diese Metriken.

## Zusammenfassung

- Velero sichert und stellt Kubernetes-Ressourcen und Volumes wieder her, gesteuert über CRDs.
- Ein Backup besteht aus Metadaten in der `BackupStorageLocation` und Volume-Daten.
- `BackupStorageLocation` ist die Single Source of Truth; `VolumeSnapshotLocation` ist nur für Cloud-Provider.
- Filesystem-Backups brauchen nur eine `BackupStorageLocation` und nutzen die VolumeSnapshotClass des external-snapshotter.
- Die Auswahl pro Volume erfolgt über Opt-in, Opt-out oder `--default-volumes-to-fs-backup`.
- `--snapshot-move-data` verschiebt Volume-Daten aus dem Filesystem-Backup nach S3.
- Die `BackupStorageLocation` kann nicht immutable sein; ein separater Bucket für Volume-Daten schon.
- Metadaten liegen unverschlüsselt, Volume-Snapshots verschlüsselt über das globale backupRepository-Secret.
- RestoreResourceModifiers liegen in einer einzigen ConfigMap, ohne Overlay-Komposition.
- Plattformübergreifende Migrationen erfordern Filesystem-Backups.
- Das Helm-Chart liefert CSI-Feature, Node-Agent, Plugins, BackupStorageLocation und Prometheus-Metriken auf Port 8085.
