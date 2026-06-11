---
title: "Provisioning a Kubernetes cluster with kubespray on Proxmox"
description: "Preparing VMs on Proxmox with cloud-init, baking the Kubernetes prerequisites into the image, running kubespray from the official container, inventory and group_vars, fetching the kubeconfig, upgrades and version pinning."
tags: [kubernetes, kubespray, proxmox, ansible, homelab]
date: 2025-03-15
---

## Overview

kubespray provisions production-ready Kubernetes clusters via Ansible. The flow has two phases: first the VMs are created on Proxmox, then kubespray rolls out Kubernetes onto them.

- Prepare a cloud-init template on Proxmox with all Kubernetes prerequisites
- Clone VMs from the template
- Run kubespray from the official container against an inventory
- Fetch the kubeconfig from the first control plane node

## VM preparation with cloud-init

kubespray expects prepared nodes. Instead of running these steps as Ansible pre-tasks, they are baked into the cloud-init template. Every cloned VM is Kubernetes-ready on first boot.

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

This disables swap (a kubelet requirement), loads the `overlay` and `br_netfilter` kernel modules and sets the sysctl parameters for bridge netfilter and IP forwarding. The qemu-guest-agent is installed as well, so Proxmox knows the VM IP.

The VMs are then created as a full clone of the template and configured via cloud-init with a static IP, an SSH key and a hostname. Only after that does kubespray come into play.

## Running kubespray from the container

kubespray is not installed locally but run from the official container image. This keeps the Ansible and Python dependencies reproducible and bound to the kubespray version.

```bash
docker run --rm -it \
  --mount type=bind,source="$(pwd)"/inventory/home-01,dst=/inventory \
  --mount type=bind,source="${HOME}"/.ssh/id_ed25519,dst=/root/.ssh/id_ed25519 \
  quay.io/kubespray/kubespray:v2.29.1 \
  ansible-playbook -i /inventory/inventory.ini \
    --private-key /root/.ssh/id_ed25519 cluster.yml
```

Two things are mounted: the inventory directory to `/inventory` and the private SSH key Ansible uses to reach the nodes. The image tag pins the kubespray version; changing the tag is a version change.

## Inventory and group_vars

The inventory lives in its own directory per cluster. It separates the hosts from the variables.

```text
inventory/home-01/
├── inventory.ini
└── group_vars/
    ├── all/            # all.yml, containerd.yml, etcd.yml
    └── k8s_cluster/    # k8s-cluster.yml, addons.yml, k8s-net-calico.yml
```

The `inventory.ini` names the nodes and their roles. The control plane and etcd group can sit on the same nodes.

```ini
[kube_control_plane]
c1 ansible_host=192.168.1.10

[etcd:children]
kube_control_plane

[kube_node]
c1 ansible_host=192.168.1.10
```

The cluster configuration sits in `group_vars/k8s_cluster/k8s-cluster.yml`.

```yaml
kube_version: v1.32.0
container_manager: containerd
kube_network_plugin: calico
kube_proxy_mode: ipvs
kube_service_addresses: 10.233.0.0/18
kube_pods_subnet: 10.233.64.0/18
```

The addons in `group_vars/k8s_cluster/addons.yml` are deliberately left disabled.

```yaml
metrics_server_enabled: false
ingress_nginx_enabled: false
local_path_provisioner_enabled: false
```

Ingress, storage and monitoring are rolled out later via GitOps, not via kubespray. This keeps the part managed by kubespray small and limited to the cluster core.

## Fetching the kubeconfig

After `cluster.yml` the admin kubeconfig is on the first control plane node at `/etc/kubernetes/admin.conf`. It points at `127.0.0.1`, though. When copying it, the loopback address is replaced by the node IP.

```bash
IP=192.168.1.10
ssh root@"$IP" "cat /etc/kubernetes/admin.conf" \
  | sed "s|https://127.0.0.1:6443|https://$IP:6443|g" \
  > inventory/home-01/kubeconfig
```

## Upgrades and maintenance

kubespray ships a dedicated playbook for each lifecycle step. They all run through the same container invocation, only the playbook at the end changes.

- `upgrade-cluster.yml` upgrades the cluster to a new version node by node in a controlled way
- `scale.yml` adds workers
- `remove-node.yml` removes a node cleanly
- `reset.yml` returns a node to its pre-kubespray state

An upgrade is a change of the image tag plus `upgrade-cluster.yml`.

```bash
docker run --rm -it \
  --mount type=bind,source="$(pwd)"/inventory/home-01,dst=/inventory \
  --mount type=bind,source="${HOME}"/.ssh/id_ed25519,dst=/root/.ssh/id_ed25519 \
  quay.io/kubespray/kubespray:v2.30.0 \
  ansible-playbook -i /inventory/inventory.ini \
    --private-key /root/.ssh/id_ed25519 upgrade-cluster.yml
```

Each kubespray version ships its own set of default `group_vars`. On a version change the local customizations must be reconciled against the new defaults. An automated three-way merge between the old and new sample version, triggered by a Renovate update PR, keeps that reconciliation traceable.

## Summary

- VMs are prepared via cloud-init; swap-off, kernel modules and sysctl are baked into the image
- kubespray runs from the official container, not installed locally
- The image tag pins the kubespray version; changing the tag is a version change
- Inventory and `group_vars` are kept separate per cluster
- `container_manager: containerd`, `kube_network_plugin: calico`, `kube_proxy_mode: ipvs`
- Addons stay disabled and are rolled out later via GitOps
- The kubeconfig is fetched from the first control plane node, the loopback address replaced
- Lifecycle through `upgrade-cluster.yml`, `scale.yml`, `remove-node.yml` and `reset.yml`
- Version changes require reconciling the local `group_vars` against the new defaults
