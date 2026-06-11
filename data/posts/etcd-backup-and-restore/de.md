---
title: "etcd sichern und wiederherstellen"
description: "etcd als Zustandsspeicher von Kubernetes, konsistente Snapshots mit etcdctl, Wiederherstellung mit etcdutl, Vorgehen für Single-Node und Multi-Node-Control-Planes, Quorum und Downtime."
tags: [etcd, kubernetes, backup, disaster-recovery]
date: 2025-10-20
---

## Warum etcd

etcd speichert den vollständigen Zustand eines Kubernetes-Clusters. Alle API-Objekte, Secrets und ConfigMaps liegen dort.

Ein konsistenter Snapshot von etcd ist die Grundlage für Disaster Recovery. Geht etcd verloren, ist das Cluster ohne Snapshot nicht rekonstruierbar.

Voraussetzung für eine Wiederherstellung: etcd und die Control-Plane-Pods müssen gestoppt sein, und die Zertifikate auf dem Zielsystem müssen zum Snapshot passen.

## Snapshot und Wiederherstellung

Zwei Werkzeuge sind beteiligt. `etcdctl` spricht mit einer laufenden etcd-Instanz und erstellt den Snapshot. `etcdutl` arbeitet offline auf einer Snapshot-Datei und stellt das Datenverzeichnis wieder her.

```bash
ETCDCTL_API=3 etcdctl ... snapshot save snapshot.db
etcdutl snapshot restore snapshot.db ...
```

Konvention: `ETCDCTL_API=3` setzt die v3-API für `etcdctl`. Der Restore läuft mit `etcdutl` ohne laufenden Server.

## Single-Node: Wiederherstellung

Vorgehen für eine Control Plane mit einem einzelnen etcd-Member.

### Snapshot erstellen

```bash
ETCDCTL_API=3 etcdctl \
  --endpoints https://127.0.0.1:2379 \
  --cacert /etc/ssl/etcd/ssl/ca.pem \
  --cert /etc/ssl/etcd/ssl/admin-node1.pem \
  --key /etc/ssl/etcd/ssl/admin-node1-key.pem \
  snapshot save snapshot.db
```

Der Snapshot wird gegen den lokalen etcd-Endpoint erstellt. Die TLS-Flags `--cacert`, `--cert` und `--key` authentisieren den Client gegenüber etcd.

### Snapshot sichern

```bash
scp <node1>:snapshot.db .
```

Die Snapshot-Datei wird vom Knoten kopiert und außerhalb des Clusters abgelegt.

### etcd-Zertifikate sichern

```bash
rsync -avz root@<node1-ip>:/etc/ssl/etcd .
```

Die Zertifikate liegen unter `/etc/ssl/etcd/ssl/`. Sie werden mitgesichert, weil der wiederhergestellte etcd-Zustand zu ihnen passen muss.

### Zielcluster provisionieren

Voraussetzung: Ein neues Cluster ist provisioniert, mit installiertem etcd und gleicher Topologie. Der Restore überschreibt anschließend dessen etcd-Datenverzeichnis.

### Control-Plane-Pods stoppen

```bash
mv /etc/kubernetes/manifests ~
```

Das Verzeichnis der statischen Pod-Manifeste wird beiseitegeschoben. Das Kubelet stoppt daraufhin `kube-apiserver`, `kube-controller-manager` und `kube-scheduler`.

### etcd stoppen

```bash
systemctl stop etcd
```

etcd muss gestoppt sein, bevor das Datenverzeichnis verändert wird.

### Aktuelles Datenverzeichnis beiseiteschieben

```bash
mv /var/lib/etcd/member /var/lib/etcd/member_bak
```

Das bestehende `member`-Verzeichnis wird umbenannt statt gelöscht. So bleibt der alte Zustand bei Bedarf erhalten.

### Snapshot wiederherstellen

```bash
etcdutl snapshot restore snapshot.db \
  --name etcd1 \
  --initial-cluster etcd1=https://<node1-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node1-ip>:2380 \
  --data-dir /var/lib/etcd/
```

`etcdutl` schreibt ein neues `member`-Verzeichnis aus dem Snapshot. `--name` und `--initial-advertise-peer-urls` beschreiben das Member, `--initial-cluster` die Cluster-Topologie.

### Zertifikate zurückspielen

```bash
rsync -avz ./etcd root@<node1-ip>:/etc/ssl/
```

Die gesicherten Zertifikate werden auf dem Zielknoten wiederhergestellt.

### etcd starten und Manifeste zurücklegen

```bash
systemctl start etcd
mv ~/manifests /etc/kubernetes/
```

etcd startet mit dem wiederhergestellten Datenverzeichnis. Danach legt das Kubelet die Control-Plane-Pods aus den Manifesten neu an.

### Status prüfen

```bash
kubectl get po -A
systemctl status etcd
journalctl -xeu etcd
```

`kubectl get po -A` zeigt, ob die API erreichbar ist und die Pods laufen. `systemctl status` und `journalctl` zeigen den etcd-Dienst und seine Logs.

## Multi-Node: Wiederherstellung

Vorgehen für eine Control Plane mit drei etcd-Membern. Der Snapshot stammt aus einem der Member und wird auf alle drei Knoten verteilt.

### Snapshot erstellen

```bash
ETCDCTL_API=3 etcdctl \
  --endpoints https://127.0.0.1:2379 \
  --cacert /etc/ssl/etcd/ssl/ca.pem \
  --cert /etc/ssl/etcd/ssl/admin-node1.pem \
  --key /etc/ssl/etcd/ssl/admin-node1-key.pem \
  snapshot save snapshot.db
```

Ein einzelner Snapshot von einem Member genügt, da alle Member denselben replizierten Zustand halten.

### Snapshot sichern

```bash
scp <node1>:snapshot.db .
```

### etcd-Zertifikate sichern

```bash
rsync -avz root@<node1-ip>:/etc/ssl/etcd .
```

Die Zertifikate liegen unter `/etc/ssl/etcd/ssl/`.

### Zielcluster provisionieren

Voraussetzung: Ein neues Cluster mit drei Control-Plane-Knoten ist provisioniert, mit installiertem etcd auf jedem Knoten.

### Snapshot an alle Control-Plane-Knoten verteilen

```bash
scp snapshot.db <node1>:.
scp snapshot.db <node2>:.
scp snapshot.db <node3>:.
```

Jeder Knoten stellt sein eigenes Member aus derselben Snapshot-Datei wieder her.

### Control-Plane-Pods auf allen Knoten stoppen

```bash
mv /etc/kubernetes/manifests ~
```

Der Schritt wird auf jedem der drei Knoten ausgeführt.

### etcd auf allen Knoten stoppen

```bash
systemctl stop etcd
```

### Aktuelles Datenverzeichnis auf allen Knoten beiseiteschieben

```bash
mv /var/lib/etcd/member ~
```

### Snapshot auf jedem Knoten wiederherstellen

Pro Member ein eigener `etcdutl`-Aufruf. `--name` und `--initial-advertise-peer-urls` unterscheiden sich je Knoten, `--initial-cluster` und `--initial-cluster-token` sind identisch.

```bash
# auf <node1>
etcdutl snapshot restore snapshot.db \
  --name etcd1 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node1-ip>:2380 \
  --data-dir /var/lib/etcd/
```

```bash
# auf <node2>
etcdutl snapshot restore snapshot.db \
  --name etcd2 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node2-ip>:2380 \
  --data-dir /var/lib/etcd/
```

```bash
# auf <node3>
etcdutl snapshot restore snapshot.db \
  --name etcd3 \
  --initial-cluster etcd1=https://<node1-ip>:2380,etcd2=https://<node2-ip>:2380,etcd3=https://<node3-ip>:2380 \
  --initial-cluster-token k8s_etcd \
  --initial-advertise-peer-urls https://<node3-ip>:2380 \
  --data-dir /var/lib/etcd/
```

Das identische `--initial-cluster` und `--initial-cluster-token` binden die drei Member zu einem Cluster zusammen. Ein abweichender Token oder eine abweichende Topologie verhindert den Zusammenschluss.

### Zertifikate auf allen Knoten zurückspielen

```bash
rsync -avz ./etcd root@<node1>:/etc/ssl/
rsync -avz ./etcd root@<node2>:/etc/ssl/
rsync -avz ./etcd root@<node3>:/etc/ssl/
```

### etcd auf allen Knoten starten

```bash
systemctl start etcd
```

Die Member starten und bilden anhand der `--initial-cluster`-Angabe das Quorum.

### Control-Plane-Pods auf allen Knoten zurücklegen

```bash
mv ~/manifests /etc/kubernetes/
```

### Status prüfen

```bash
kubectl get po -A
systemctl status etcd
journalctl -xeu etcd
```

## Quorum und Downtime

Ein etcd-Cluster mit drei Membern toleriert den Ausfall eines Members. Schreibzugriffe benötigen die Mehrheit, also zwei der drei Member.

Während des Restores sind alle Member gestoppt. Das Cluster ist in dieser Phase nicht verfügbar; die Downtime umfasst den gesamten Ablauf bis zum erneuten Erreichen des Quorums.

Konvention: Die Member erst starten, wenn alle drei Datenverzeichnisse aus demselben Snapshot wiederhergestellt sind. Andernfalls entsteht kein konsistentes Quorum.

## Zusammenfassung

- etcd hält den gesamten Cluster-Zustand; ohne Snapshot ist keine Wiederherstellung möglich
- `etcdctl` erstellt den Snapshot online, `etcdutl` stellt ihn offline wieder her
- `ETCDCTL_API=3` und die TLS-Flags `--cacert`, `--cert`, `--key` sind beim Snapshot erforderlich
- Vor dem Restore Control-Plane-Manifeste beiseiteschieben und etcd stoppen
- Das bestehende `member`-Verzeichnis umbenennen, nicht löschen
- Zertifikate müssen zum wiederhergestellten Zustand passen
- Single-Node: ein `etcdutl`-Aufruf; Multi-Node: ein Aufruf pro Member
- Multi-Node teilt sich `--initial-cluster` und `--initial-cluster-token`, unterscheidet `--name` und `--initial-advertise-peer-urls`
- Drei Member tolerieren einen Ausfall; das Quorum sind zwei Member
- Während des Restores ist das Cluster nicht verfügbar
