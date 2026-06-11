---
title: "Velero: filesystem backups and S3 export"
description: "Structure of a Velero backup, BackupStorageLocation and VolumeSnapshotLocation, filesystem backups, exporting volume data to S3 with snapshot-move-data, immutability, encryption, RestoreResourceModifiers, migrations, Helm setup and monitoring."
tags: [velero, kubernetes, backup, s3]
date: 2025-09-15
---

## Structure

Velero consists of a server/controller in the cluster and a CLI. Its purpose is backup and restore of Kubernetes resources and volumes.

```bash
velero backup create my-backup
```

Velero works on the basis of CRDs. Backups, restores and schedules are therefore custom resources in the cluster. Backups run on-demand or on a schedule and carry an optional TTL for expiration and retention.

A Velero backup consists of two parts:

- Metadata and manifests in an object store (single source of truth)
- Volume data, either through a cloud provider with VolumeSnapshot support or through a filesystem backup

## Backup procedure

The procedure is described in the Velero documentation under `how-velero-works`.

```bash
velero backup create my-backup --include-namespaces app
```

Procedure: the CLI creates a backup CR in the current Kubernetes context. The BackupController detects the CR, validates it and fetches the manifests from the Kubernetes API. The manifests are uploaded to the `BackupStorageLocation`. For PersistentVolumes the controller selects the matching location depending on the configuration and stores the volume data there. Volume snapshots can be disabled with `--snapshot-volumes=false`.

## Backup and volume locations

Since a backup splits into two parts, there are two custom resources for the locations.

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

`BackupStorageLocation`: an S3 bucket holding all Velero data such as YAML and metadata. Single source of truth.

`VolumeSnapshotLocation`: the place where volume snapshots are stored. The metadata in the `BackupStorageLocation` references this location. This location is intended only for cloud providers. In a cloud cluster the provider supplies its own snapshot classes, and a provider-specific plugin makes them usable in Velero.

Filesystem backups do not require a `VolumeSnapshotLocation`. There Velero uses the Kubernetes-internal VolumeSnapshotClasses.

## Filesystem backups

A filesystem backup requires only a `BackupStorageLocation` and no `VolumeSnapshotLocation`. The metadata resides in the S3 object store, while the volume data is backed up through the filesystem, for example through Ceph.

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

For this Velero uses a `VolumeSnapshotClass` of the external-snapshotter (`snapshot.storage.k8s.io`). The label `velero.io/csi-volumesnapshot-class: "true"` marks the class relevant to Velero.

Convention: if no `VolumeSnapshotLocation` is defined and filesystem backups are configured, Velero uses filesystem backups by default, since nothing else is available. The control options below are irrelevant in this case.

### Opt-in

If both a `VolumeSnapshotLocation` and filesystem backups are enabled, an annotation controls the choice per volume.

```yaml
metadata:
  annotations:
    backup.velero.io/backup-volumes: pvc-volume,emptydir-volume
```

By default Velero performs a filesystem backup only if the Kubernetes object carries this annotation. Volumes without the annotation are considered for regular VolumeSnapshots if `--snapshot-volumes` is not set to `false` and the volume is snapshot-capable.

From this follows:

- Volumes with and without the annotation, `--snapshot-volumes=false`: filesystem backups only.
- Volumes with and without the annotation, `--snapshot-volumes=true` (default): filesystem backup for annotated volumes, VolumeSnapshots for the rest.

### Opt-out

```yaml
metadata:
  annotations:
    backup.velero.io/backup-volumes-excludes: log-volume
```

With `backup.velero.io/backup-volumes-excludes` individual volumes are excluded from the filesystem backup.

### Default to filesystem backup

```bash
velero install --default-volumes-to-fs-backup
```

With `--default-volumes-to-fs-backup=true` the default is always a filesystem backup. If a backup targets a cloud provider, the `VolumeSnapshotLocation` then has to be set explicitly.

## Exporting volume data to S3

An S3-compatible store without a Velero volume snapshot plugin cannot take volume data directly. Velero backs up volumes either through a `VolumeSnapshotLocation` or through filesystem backups.

Problem: if a provider supplies only S3 and has no dedicated snapshot plugin for Kubernetes, there is no matching Velero provider plugin for the volume snapshots.

Solution: volume data from a filesystem backup can be moved into an object store after the backup.

```bash
velero backup create my-backup --snapshot-move-data
```

The backup is created with `--snapshot-move-data`. By default the volume data lands in the same bucket as the `BackupStorageLocation`. A custom data mover defines the target more precisely.

```bash
velero backup create my-backup --snapshot-move-data --data-mover <data-mover-name>
```

This way the AWS plugin configures the S3 store as the `BackupStorageLocation`, and through filesystem backup plus `snapshot-move-data` the volume data is pushed to S3.

## OpenStack provider

For cloud clusters on OpenStack the `velero-plugin-for-openstack` can be used, which drives the Cinder volume snapshotter.

```yaml
initContainers:
  - name: velero-plugin-for-openstack
    image: velero/velero-plugin-for-openstack:v0.8.0
    volumeMounts:
      - mountPath: /target
        name: plugins
```

In this case the provider's S3 serves as the `BackupStorageLocation`, and Cinder supplies the `VolumeSnapshotLocation`.

## RestoreResourceModifiers

RestoreResourceModifiers manipulate Kubernetes resources before the restore. They reside in a ConfigMap and are referenced at restore time.

```bash
velero restore create --from-backup my-backup --resource-modifier-configmap my-modifiers
```

A restore can reference only a single ConfigMap. All transformations must be in this ConfigMap. An overlay principle that composes several sets is not possible.

## Migrations

A complete backup and restore of a cluster is called a migration.

```bash
velero backup create cluster-full
velero restore create --from-backup cluster-full
```

Natively this works only when both clusters run on the same cloud platform. A migration between different platforms requires filesystem backups.

## Immutability

From version 1.11 on, Velero backups do not work reliably when the target object store is configured with immutability.

```text
Cannot support backup data immutability
```

This concerns the `BackupStorageLocation`. It is the single source of truth into which Velero continuously writes the status of backups. No immutability can be configured for this bucket.

A separate bucket for the exported volume data can, however, be immutable. The only consequence would be that Velero can delete a volume only after the immutability period has elapsed. If the backup TTL is set equal to the bucket's immutability period, this problem does not arise.

## Encryption

The metadata and YAML manifests reside unencrypted in the `BackupStorageLocation`. The volume snapshots are encrypted.

```yaml
credentials:
  existingSecret: velero-repo-credentials
```

The key resides in a global secret deployed by the Helm chart. This backupRepository secret encrypts all backupRepositories.

## Helm setup

Velero can be deployed through a Helm chart.

```bash
helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts/
helm install velero vmware-tanzu/velero --namespace velero --values values.yaml
```

Required configuration for filesystem backups in the Helm values:

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

`features: EnableCSI` enables CSI support, `deployNodeAgent: true` starts the node agent for filesystem backups. The toleration allows backups of workloads on the control plane. The VolumeSnapshotClasses in use need the label `velero.io/csi-volumesnapshot-class: "true"`.

The provider plugins are added as init containers:

```yaml
initContainers:
  - name: velero-plugin-for-aws
    image: velero/velero-plugin-for-aws:v1.12.2
    imagePullPolicy: IfNotPresent
    volumeMounts:
      - mountPath: /target
        name: plugins
```

The `BackupStorageLocation` is configured under `configuration.backupStorageLocation`:

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

`s3ForcePathStyle: true` switches the AWS SDK address style to path-style, which S3-compatible stores such as MinIO expect. For the OpenStack scenario a `VolumeSnapshotLocation` is additionally set under `configuration.volumeSnapshotLocation`.

## Monitoring

Velero exposes Prometheus metrics on port 8085.

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

The endpoint is scraped through a ServiceMonitor. Based on the metrics, Alertmanager can send notifications. Velero maintains its own Grafana dashboard for these metrics.

## Summary

- Velero backs up and restores Kubernetes resources and volumes, driven by CRDs.
- A backup consists of metadata in the `BackupStorageLocation` and volume data.
- `BackupStorageLocation` is the single source of truth; `VolumeSnapshotLocation` is only for cloud providers.
- Filesystem backups need only a `BackupStorageLocation` and use the external-snapshotter VolumeSnapshotClass.
- The per-volume choice is made through opt-in, opt-out or `--default-volumes-to-fs-backup`.
- `--snapshot-move-data` moves volume data from the filesystem backup to S3.
- The `BackupStorageLocation` cannot be immutable; a separate bucket for volume data can.
- Metadata resides unencrypted, volume snapshots encrypted through the global backupRepository secret.
- RestoreResourceModifiers reside in a single ConfigMap, without overlay composition.
- Cross-platform migrations require filesystem backups.
- The Helm chart provides the CSI feature, node agent, plugins, BackupStorageLocation and Prometheus metrics on port 8085.
