---
title: "Kubernetes-Cluster mit kubespray auf Proxmox provisionieren"
description: "VMs auf Proxmox per cloud-init vorbereiten, Kubernetes-Voraussetzungen ins Image backen, kubespray aus dem offiziellen Container ausführen, Inventory und group_vars, kubeconfig holen, Upgrades und Versions-Pinning."
tags: [kubernetes, kubespray, proxmox, ansible, homelab]
date: 2025-03-15
---

## Überblick

kubespray provisioniert produktionsreife Kubernetes-Cluster über Ansible. Der Ablauf gliedert sich in zwei Phasen: zuerst entstehen die VMs auf Proxmox, dann rollt kubespray Kubernetes darauf aus.

- Cloud-init-Template auf Proxmox vorbereiten, mit allen Kubernetes-Voraussetzungen
- VMs aus dem Template klonen
- kubespray aus dem offiziellen Container gegen ein Inventory ausführen
- kubeconfig vom ersten Control-Plane-Knoten holen

## VM-Vorbereitung mit cloud-init

kubespray erwartet vorbereitete Knoten. Statt diese Schritte als Ansible-Pre-Tasks laufen zu lassen, werden sie in das cloud-init-Template gebacken. Jede geklonte VM ist damit beim ersten Start Kubernetes-ready.

```yaml
#cloud-config
runcmd:
  - swapoff -a
  - sed -i '/ swap / s/^/#/' /etc/fstab
  - |
    cat > /etc/modules-load.d/k8s.conf << EOF
    overlay
    br_netfilter
    EOF
  - modprobe overlay
  - modprobe br_netfilter
  - |
    cat > /etc/sysctl.d/99-kubernetes.conf << EOF
    net.bridge.bridge-nf-call-iptables = 1
    net.bridge.bridge-nf-call-ip6tables = 1
    net.ipv4.ip_forward = 1
    fs.inotify.max_user_instances = 8192
    fs.inotify.max_user_watches = 524288
    vm.max_map_count = 262144
    EOF
  - sysctl --system
```

Das deaktiviert Swap (Anforderung des kubelet), lädt die Kernel-Module `overlay` und `br_netfilter` und setzt die sysctl-Parameter für Bridge-Netfilter und IP-Forwarding. Der qemu-guest-agent wird ebenfalls installiert, damit Proxmox die VM-IP kennt.

Die VMs werden anschließend als Full-Clone aus dem Template erzeugt und per cloud-init mit statischer IP, SSH-Key und Hostname konfiguriert. Erst danach kommt kubespray ins Spiel.

## kubespray aus dem Container ausführen

kubespray wird nicht lokal installiert, sondern aus dem offiziellen Container-Image ausgeführt. Das hält die Ansible- und Python-Abhängigkeiten reproduzierbar und an die kubespray-Version gebunden.

```bash
docker run --rm -it \
  --mount type=bind,source="$(pwd)"/inventory/home-01,dst=/inventory \
  --mount type=bind,source="${HOME}"/.ssh/id_ed25519,dst=/root/.ssh/id_ed25519 \
  quay.io/kubespray/kubespray:v2.29.1 \
  ansible-playbook -i /inventory/inventory.ini \
    --private-key /root/.ssh/id_ed25519 cluster.yml
```

Gemountet werden zwei Dinge: das Inventory-Verzeichnis nach `/inventory` und der private SSH-Key, mit dem Ansible die Knoten erreicht. Das Image-Tag legt die kubespray-Version fest; ein Wechsel des Tags ist ein Versions-Wechsel.

## Inventory und group_vars

Das Inventory liegt pro Cluster in einem eigenen Verzeichnis. Es trennt die Hosts von den Variablen.

```text
inventory/home-01/
├── inventory.ini
└── group_vars/
    ├── all/            # all.yml, containerd.yml, etcd.yml
    └── k8s_cluster/    # k8s-cluster.yml, addons.yml, k8s-net-calico.yml
```

Die `inventory.ini` benennt die Knoten und ihre Rollen. Control-Plane- und etcd-Gruppe können auf denselben Knoten liegen.

```ini
[kube_control_plane]
c1 ansible_host=192.168.1.10

[etcd:children]
kube_control_plane

[kube_node]
c1 ansible_host=192.168.1.10
```

Die Cluster-Konfiguration steht in `group_vars/k8s_cluster/k8s-cluster.yml`.

```yaml
kube_version: v1.32.0
container_manager: containerd
kube_network_plugin: calico
kube_proxy_mode: ipvs
kube_service_addresses: 10.233.0.0/18
kube_pods_subnet: 10.233.64.0/18
```

Die Addons in `group_vars/k8s_cluster/addons.yml` bleiben bewusst deaktiviert.

```yaml
metrics_server_enabled: false
ingress_nginx_enabled: false
local_path_provisioner_enabled: false
```

Ingress, Storage und Monitoring werden später über GitOps ausgerollt, nicht über kubespray. Damit bleibt der von kubespray verwaltete Teil klein und auf den Cluster-Kern beschränkt.

## kubeconfig holen

Nach `cluster.yml` liegt die Admin-kubeconfig auf dem ersten Control-Plane-Knoten unter `/etc/kubernetes/admin.conf`. Sie zeigt jedoch auf `127.0.0.1`. Beim Kopieren wird die Loopback-Adresse durch die Knoten-IP ersetzt.

```bash
IP=192.168.1.10
ssh root@"$IP" "cat /etc/kubernetes/admin.conf" \
  | sed "s|https://127.0.0.1:6443|https://$IP:6443|g" \
  > inventory/home-01/kubeconfig
```

## Upgrades und Wartung

kubespray bringt für jeden Lebenszyklus-Schritt ein eigenes Playbook mit. Alle laufen über dieselbe Container-Invocation, nur das Playbook am Ende wechselt.

- `upgrade-cluster.yml` hebt das Cluster kontrolliert Knoten für Knoten auf eine neue Version
- `scale.yml` fügt Worker hinzu
- `remove-node.yml` entfernt einen Knoten sauber
- `reset.yml` setzt einen Knoten auf den Zustand vor kubespray zurück

Ein Upgrade ist ein Wechsel des Image-Tags plus `upgrade-cluster.yml`.

```bash
docker run --rm -it \
  --mount type=bind,source="$(pwd)"/inventory/home-01,dst=/inventory \
  --mount type=bind,source="${HOME}"/.ssh/id_ed25519,dst=/root/.ssh/id_ed25519 \
  quay.io/kubespray/kubespray:v2.30.0 \
  ansible-playbook -i /inventory/inventory.ini \
    --private-key /root/.ssh/id_ed25519 upgrade-cluster.yml
```

Jede kubespray-Version bringt ein eigenes Set an Default-`group_vars` mit. Beim Versions-Wechsel müssen die lokalen Anpassungen gegen die neuen Defaults abgeglichen werden. Ein automatisierter Drei-Wege-Merge zwischen alter und neuer Sample-Version, ausgelöst durch einen Renovate-Update-PR, hält diesen Abgleich nachvollziehbar.

## Zusammenfassung

- VMs werden per cloud-init vorbereitet; Swap-off, Kernel-Module und sysctl sind ins Image gebacken
- kubespray läuft aus dem offiziellen Container, nicht lokal installiert
- Das Image-Tag bindet die kubespray-Version; ein Tag-Wechsel ist ein Versions-Wechsel
- Inventory und `group_vars` liegen pro Cluster getrennt
- `container_manager: containerd`, `kube_network_plugin: calico`, `kube_proxy_mode: ipvs`
- Addons bleiben deaktiviert und werden später über GitOps ausgerollt
- Die kubeconfig wird vom ersten Control-Plane-Knoten geholt, die Loopback-Adresse ersetzt
- Lifecycle über `upgrade-cluster.yml`, `scale.yml`, `remove-node.yml` und `reset.yml`
- Versions-Wechsel erfordern einen Abgleich der lokalen `group_vars` gegen die neuen Defaults
