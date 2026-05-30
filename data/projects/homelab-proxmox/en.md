---
title: "Homelab: Proxmox"
description: "Single-node Proxmox homelab provisioned end to end with Ansible. From a preseeded Debian installer to running Kubernetes clusters: ZFS for storage, cloud-init for VM templates, Kubespray for the clusters."
tags: [ansible, proxmox, kubernetes, kubespray, zfs, cloud-init]
date: 2026-05-30
repo: https://github.com/srkn0/homelab-proxmox
---

## Overview

Infrastructure as Code for my Proxmox node, from bare metal to a running Kubernetes cluster. Each step is its own playbook or role, all idempotent.

## Credits

Proxmox setup uses the [lae.proxmox](https://github.com/lae/ansible-role-proxmox) role, ZFS the [mrlesmithjr.zfs](https://github.com/mrlesmithjr/ansible-zfs) role, both pinned in `requirements.yml` and installed via ansible-galaxy. Clusters are bootstrapped with [Kubespray](https://github.com/kubernetes-sigs/kubespray). The preseed ISO build technique (repacking the initrd via cpio, remastering with xorriso) I took from [Paul Lockaby's preseed project](https://github.com/paullockaby/debian-preseed) and packaged into an Ansible role.

## Stack & Architecture

- Ansible for host and VM provisioning
- Proxmox VE as the hypervisor
- ZFS for storage
- Debian preseed for the bare-metal installer
- cloud-init with Ubuntu cloud images for the VM templates
- Kubespray for Kubernetes
- go-task as the task runner, managed via mise
- yamllint and ansible-lint, via GitHub Actions and pre-commit
- Renovate for automated dependency updates

**Bare metal:** The `debian_preseed_iso` role remasters a Debian netinst ISO into a preseeded installer. Locale, networking, root account and SSH keys come from variables; the role renders the config and wraps the build script. The result is a root-only host with a `vmbr0` bridge, ready for Proxmox.

**Host:** ZFS pools and Proxmox are configured by two playbooks. ZFS first, then Proxmox VE on top, including PCIe/IOMMU GPU passthrough for the RTX 3060 (OVMF, vfio, driver blacklist).

**Templates:** The `proxmox_cloudinit_template` role downloads an Ubuntu cloud image, verifies its SHA-256 checksum, resizes it and wraps it in a VM with a cloud-init vendor snippet, then converts it to a template. Three templates exist: a generic one, a Kubernetes-ready one (swap off, kernel modules, sysctls) and an NVIDIA/AI dev one.

**VMs:** The `proxmox_vm` role full-clones a template, sets sizing, resizes the disk and applies cloud-init network and user config. It provisions the Kubernetes nodes and the dev VM.

**Clusters:** A Taskfile drives Kubespray inside the official container image. One inventory per cluster under `inventory/kubespray/`. Targets bootstrap a cluster, upgrade it to a given Kubernetes version, and fetch the kubeconfig from the first control plane.

## Storage

One pool layout, three roles. `zpool_k8s` is four Samsung SATA SSDs as two mirror vdevs, striped (Kubernetes workloads). `zpool_nvme` is a single 1 TB NVMe for VM disks. `zpool_backup` is a 2 TB HDD. Per-dataset properties (compression, recordsize, atime) and ARC limits are set in the playbook.

## Dependency automation

Renovate (`.renovaterc.json5`) keeps the external roles, the GitHub Actions, the pre-commit hooks and the Kubespray version current. Bumps get context right in the PR. For Kubespray a workflow 3-way merges the upstream sample delta into the customized `group_vars` and posts the release notes for the versions in between. For the `lae.proxmox` and `mrlesmithjr.zfs` roles a workflow comments the `defaults/main.yml` diff and flags which of the variables I set are removed, renamed, or have a changed default.

## Planned

- Actually using the RTX 3060: GPU passthrough into a workload VM
- Declarative backups for the ZFS pools
