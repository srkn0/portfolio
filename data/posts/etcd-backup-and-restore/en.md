---
title: "Backing up and restoring etcd"
description: "etcd as the state store of Kubernetes, consistent snapshots with etcdctl, restore with etcdutl, procedures for single-node and multi-node control planes, quorum and downtime."
tags: [etcd, kubernetes, backup, disaster-recovery]
date: 2025-10-20
---

## Why etcd

etcd stores the complete state of a Kubernetes cluster. All API objects, secrets and ConfigMaps reside there.

A consistent snapshot of etcd is the basis for disaster recovery. If etcd is lost, the cluster cannot be reconstructed without a snapshot.

Prerequisite for a restore: etcd and the control-plane pods must be stopped, and the certificates on the target system must match the snapshot.

## Snapshot and restore

Two tools are involved. `etcdctl` talks to a running etcd instance and creates the snapshot. `etcdutl` works offline on a snapshot file and restores the data directory.

```bash
ETCDCTL_API=3 etcdctl ... snapshot save snapshot.db
etcdutl snapshot restore snapshot.db ...
```

Convention: `ETCDCTL_API=3` selects the v3 API for `etcdctl`. The restore runs with `etcdutl` without a running server.

## Single-node: restore

Procedure for a control plane with a single etcd member.

### Create the snapshot

```bash
ETCDCTL_API=3 etcdctl \
  --endpoints https://127.0.0.1:2379 \
  --cacert /etc/ssl/etcd/ssl/ca.pem \
  --cert /etc/ssl/etcd/ssl/admin-node1.pem \
  --key /etc/ssl/etcd/ssl/admin-node1-key.pem \
  snapshot save snapshot.db
```

The snapshot is taken against the local etcd endpoint. The TLS flags `--cacert`, `--cert` and `--key` authenticate the client to etcd.

### Copy the snapshot

```bash
scp <node1>:snapshot.db .
```

The snapshot file is copied from the node and stored outside the cluster.

### Back up the etcd certificates

```bash
rsync -avz root@<node1-ip>:/etc/ssl/etcd .
```

The certificates reside under `/etc/ssl/etcd/ssl/`. They are backed up because the restored etcd state must match them.

### Provision the target cluster

Prerequisite: A new cluster is provisioned, with etcd installed and the same topology. The restore then overwrites its etcd data directory.

### Stop the control-plane pods

```bash
mv /etc/kubernetes/manifests ~
```

The directory of static pod manifests is moved aside. The kubelet then stops `kube-apiserver`, `kube-controller-manager` and `kube-scheduler`.

### Stop etcd

```bash
systemctl stop etcd
```

etcd must be stopped before the data directory is modified.

### Move the current data directory aside

```bash
mv /var/lib/etcd/member /var/lib/etcd/member_bak
```

The existing `member` directory is renamed rather than deleted. The old state is kept in case it is needed.

### Restore the snapshot

```bash
etcdutl snapshot restore snapshot.db \
  --name etcd1 \
  --initial-cluster etcd1=https://<node1-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node1-ip>:2380 \
  --data-dir /var/lib/etcd/
```

`etcdutl` writes a new `member` directory from the snapshot. `--name` and `--initial-advertise-peer-urls` describe the member, `--initial-cluster` the cluster topology.

### Restore the certificates

```bash
rsync -avz ./etcd root@<node1-ip>:/etc/ssl/
```

The backed-up certificates are restored on the target node.

### Start etcd and move the manifests back

```bash
systemctl start etcd
mv ~/manifests /etc/kubernetes/
```

etcd starts with the restored data directory. The kubelet then recreates the control-plane pods from the manifests.

### Verify the status

```bash
kubectl get po -A
systemctl status etcd
journalctl -xeu etcd
```

`kubectl get po -A` shows whether the API is reachable and the pods are running. `systemctl status` and `journalctl` show the etcd service and its logs.

## Multi-node: restore

Procedure for a control plane with three etcd members. The snapshot comes from one of the members and is distributed to all three nodes.

### Create the snapshot

```bash
ETCDCTL_API=3 etcdctl \
  --endpoints https://127.0.0.1:2379 \
  --cacert /etc/ssl/etcd/ssl/ca.pem \
  --cert /etc/ssl/etcd/ssl/admin-node1.pem \
  --key /etc/ssl/etcd/ssl/admin-node1-key.pem \
  snapshot save snapshot.db
```

A single snapshot from one member is sufficient, since all members hold the same replicated state.

### Copy the snapshot

```bash
scp <node1>:snapshot.db .
```

### Back up the etcd certificates

```bash
rsync -avz root@<node1-ip>:/etc/ssl/etcd .
```

The certificates reside under `/etc/ssl/etcd/ssl/`.

### Provision the target cluster

Prerequisite: A new cluster with three control-plane nodes is provisioned, with etcd installed on each node.

### Distribute the snapshot to all control-plane nodes

```bash
scp snapshot.db <node1>:.
scp snapshot.db <node2>:.
scp snapshot.db <node3>:.
```

Each node restores its own member from the same snapshot file.

### Stop the control-plane pods on all nodes

```bash
mv /etc/kubernetes/manifests ~
```

This step is run on each of the three nodes.

### Stop etcd on all nodes

```bash
systemctl stop etcd
```

### Move the current data directory aside on all nodes

```bash
mv /var/lib/etcd/member ~
```

### Restore the snapshot on each node

One `etcdutl` invocation per member. `--name` and `--initial-advertise-peer-urls` differ per node, `--initial-cluster` and `--initial-cluster-token` are identical.

```bash
# on <node1>
etcdutl snapshot restore snapshot.db \
  --name etcd1 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node1-ip>:2380 \
  --data-dir /var/lib/etcd/
```

```bash
# on <node2>
etcdutl snapshot restore snapshot.db \
  --name etcd2 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node2-ip>:2380 \
  --data-dir /var/lib/etcd/
```

```bash
# on <node3>
etcdutl snapshot restore snapshot.db \
  --name etcd3 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node3-ip>:2380 \
  --data-dir /var/lib/etcd/
```

The identical `--initial-cluster` and `--initial-cluster-token` bind the three members into one cluster. A diverging token or topology prevents them from joining.

### Restore the certificates on all nodes

```bash
rsync -avz ./etcd root@<node1>:/etc/ssl/
rsync -avz ./etcd root@<node2>:/etc/ssl/
rsync -avz ./etcd root@<node3>:/etc/ssl/
```

### Start etcd on all nodes

```bash
systemctl start etcd
```

The members start and form the quorum based on the `--initial-cluster` setting.

### Move the control-plane pods back on all nodes

```bash
mv ~/manifests /etc/kubernetes/
```

### Verify the status

```bash
kubectl get po -A
systemctl status etcd
journalctl -xeu etcd
```

## Quorum and downtime

An etcd cluster with three members tolerates the loss of one member. Writes require the majority, that is two of the three members.

During the restore all members are stopped. The cluster is unavailable in this phase; the downtime spans the entire procedure until quorum is reached again.

Convention: Start the members only after all three data directories have been restored from the same snapshot. Otherwise no consistent quorum forms.

## Summary

- etcd holds the entire cluster state; without a snapshot no restore is possible
- `etcdctl` creates the snapshot online, `etcdutl` restores it offline
- `ETCDCTL_API=3` and the TLS flags `--cacert`, `--cert`, `--key` are required for the snapshot
- Before the restore, move the control-plane manifests aside and stop etcd
- Rename the existing `member` directory, do not delete it
- Certificates must match the restored state
- Single-node: one `etcdutl` invocation; multi-node: one invocation per member
- Multi-node shares `--initial-cluster` and `--initial-cluster-token`, differs in `--name` and `--initial-advertise-peer-urls`
- Three members tolerate one failure; the quorum is two members
- The cluster is unavailable during the restore
