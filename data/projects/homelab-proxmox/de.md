---
title: "Homelab: Proxmox"
description: "Single-Node-Proxmox-Homelab, komplett mit Ansible provisioniert. Vom preseedeten Debian-Installer bis zu laufenden Kubernetes-Clustern: ZFS für Storage, cloud-init für VM-Templates, Kubespray für die Cluster."
tags: [ansible, proxmox, kubernetes, kubespray, zfs, cloud-init]
date: 2026-05-30
category: infrastructure
featured: 2
repo: https://github.com/srkn0/homelab-proxmox
---

## Überblick

Infrastructure as Code für meine Proxmox Node, von der nackten Maschine bis zum laufenden Kubernetes-Cluster. Jeder Schritt ist ein eigenes Playbook oder eine Role, alles idempotent.

## Credits

Das Proxmox-Setup nutzt die Role [lae.proxmox](https://github.com/lae/ansible-role-proxmox), ZFS die Role [mrlesmithjr.zfs](https://github.com/mrlesmithjr/ansible-zfs), beide in `requirements.yml` gepinnt und über ansible-galaxy installiert. Die Cluster werden mit [Kubespray](https://github.com/kubernetes-sigs/kubespray) bootstrapped. Die Build-Technik für das Preseed-ISO (Initrd per cpio neu packen, mit xorriso re-mastern) habe ich von [Paul Lockabys Preseed-Projekt](https://github.com/paullockaby/debian-preseed) übernommen und in eine Ansible-Rolle gepackt.

## Stack & Architektur

- Ansible für Host- und VM-Provisioning
- Proxmox VE als Hypervisor
- ZFS als Storage
- Debian Preseed für den Bare-Metal-Installer
- cloud-init mit Ubuntu-Cloud-Images für die VM-Templates
- Kubespray für Kubernetes
- go-task als Task-Runner, verwaltet über mise
- yamllint und ansible-lint, über GitHub Actions und pre-commit
- Renovate für automatisierte Dependency-Updates

**Bare Metal:** Die Role `debian_preseed_iso` baut aus einem Debian-Netinst-ISO einen preseedeten Installer. Locale, Netzwerk, Root-Account und SSH-Keys kommen aus Variablen; die Role rendert die Config und kapselt das Build-Skript. Ergebnis ist ein Root-only-Host mit `vmbr0`-Bridge, fertig für Proxmox.

**Host:** ZFS-Pools und Proxmox werden über zwei Playbooks konfiguriert. Erst ZFS, dann Proxmox VE darauf, inklusive PCIe/IOMMU-GPU-Passthrough für die RTX 3060 (OVMF, vfio, Treiber-Blacklist).

**Templates:** Die Role `proxmox_cloudinit_template` lädt ein Ubuntu-Cloud-Image, prüft die SHA-256-Checksumme, resized es und kapselt es in eine VM mit cloud-init-Vendor-Snippet, dann wird daraus ein Template. Es gibt drei Templates: ein generisches, ein Kubernetes-fertiges (Swap aus, Kernel-Module, Sysctls) und ein NVIDIA/AI-Dev-Template.

**VMs:** Die Role `proxmox_vm` klont ein Template (Full Clone), setzt das Sizing, resized die Disk und wendet cloud-init-Netzwerk- und User-Config an. Damit werden die Kubernetes-Nodes und die Dev-VM provisioniert.

**Cluster:** Ein Taskfile steuert Kubespray im offiziellen Container-Image. Ein Inventory pro Cluster unter `inventory/kubespray/`. Es gibt Targets zum bootstrappen der Cluster, für Upgrades und um die kubeconfig von der ersten Control-Plane-Node zu holen.

## Storage

Ein Pool-Layout, drei Rollen. `zpool_k8s` sind vier Samsung-SATA-SSDs als zwei Mirror-Vdevs, gestriped (für Kubernetes-PVs, per Proxmox-csi). `zpool_nvme` ist eine einzelne 1-TB-NVMe für die VM-Disks. `zpool_backup` ist eine 2-TB-HDD. Properties pro Dataset (Compression, Recordsize, atime) und ARC-Limits werden im Playbook gesetzt.

## Dependency-Automatisierung

Renovate (`.renovaterc.json5`) hält die externen Roles, die GitHub Actions, die pre-commit-Hooks und die Kubespray-Version aktuell. Die Bumps bekommen Kontext direkt im PR. Bei Kubespray merged ein Workflow den Upstream-Sample-Diff per 3-Wege-Merge in die customizten `group_vars` und postet die Release-Notes der Versionen dazwischen. Bei den Roles `lae.proxmox` und `mrlesmithjr.zfs` kommentiert ein Workflow den `defaults/main.yml`-Diff und markiert, welche der von mir gesetzten Variablen entfernt, umbenannt oder mit geändertem Default kommen.

## Geplant

- Die RTX 3060 tatsächlich nutzen: GPU-Passthrough in eine Workload-VM
- Deklarative Backups für die ZFS-Pools
