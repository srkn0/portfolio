---
title: "Provisioning Proxmox and ZFS declaratively"
description: "Bare-metal Proxmox with Ansible: a preseeded Debian image as the installer, Proxmox VE with IOMMU GPU passthrough, ZFS pools over stable by-id devices, ARC tuning and ZED alerting, Proxmox storage on ZFS, a reproducible flow from the bare machine to VM storage."
tags: [proxmox, zfs, ansible, debian, homelab]
date: 2025-02-10
---

## Overview

A Proxmox host can be built entirely declaratively, from the bare machine to usable VM storage. Every step is an Ansible playbook or role, and every step is idempotent.

The flow has four stages: a preseeded Debian image installs the base, Ansible sets up Proxmox VE, ZFS pools are created over stable device names, and Proxmox receives those pools as storage. The full setup lives in the [homelab-proxmox](/projects/homelab-proxmox) project.

## Debian preseed image

The bare-metal entry point is a preseeded Debian netinst image. An Ansible role remasters the official ISO: it renders a `preseed.cfg` and a post-install script and packs both into the initrd via `cpio`, then rewrites the ISO with `xorriso`.

```bash
ansible-playbook playbooks/build_preseed_iso.yml \
  -e input_iso=debian-13-amd64-netinst.iso \
  -e output_iso=preseed-debian-13-amd64-netinst.iso
```

The `preseed.cfg` answers locale, timezone, NTP, package selection and the SSH server. Partitioning stays interactive on purpose, so the target medium is confirmed manually and no disk is overwritten by accident.

The post-install script brings the host into the state Proxmox expects. It replaces `systemd-timesyncd` with `chrony`, creates the `vmbr0` bridge and installs the SSH key for root access.

```text
auto vmbr0
iface vmbr0 inet static
    address 192.168.1.10/24
    gateway 192.168.1.1
    bridge-ports <nic>
    bridge-stp off
    bridge-fd 0
```

The root password and `authorized_keys` come from variables. The result is a minimal Debian host with a ready bridge, reachable by SSH key, prepared for Proxmox.

## Proxmox VE

Proxmox is set up through the `lae.proxmox` role, pinned to a version in `requirements.yml` and installed via ansible-galaxy. The role sets the no-subscription repository, removes the subscription notice and installs `proxmox-ve`.

For GPU passthrough the same role enables IOMMU and binds the card to `vfio-pci`. In the homelab this is an RTX 3060, intended to be passed through to a workload VM later.

```yaml
pve_pcie_passthrough_enabled: true
pve_iommu_passthrough_mode: true
pve_pci_device_ids:
  - id: "<gpu-pci-id>"
pve_vfio_blacklist_drivers:
  - nouveau
  - nvidia
```

This sets `intel_iommu=on iommu=pt` in GRUB, blacklists the host drivers and loads the `vfio` modules. The card then belongs to the hypervisor and can be bound exclusively to a VM.

## ZFS pools

The pools are created through the `mrlesmithjr.zfs` role, driven by a list in the playbook. Devices are referenced by their `/dev/disk/by-id/` name, not by `sdX`. The by-id name stays stable across reboots and slot changes; a pool made of `sdb` and `sdc` would be a gamble after the next boot.

```yaml
zfs_pools:
  - name: zpool_k8s
    type: mirror
    devices:
      - ata-Samsung_SSD_870_EVO_500GB_<serial-a>
      - ata-Samsung_SSD_870_EVO_500GB_<serial-b>
    compression: lz4
    recordsize: 16k
    atime: "off"
    mountpoint: /mnt/zpool_k8s
```

One pool layout, three roles. `zpool_k8s` is SATA SSDs as mirror vdevs, striped, for Kubernetes volumes. `zpool_nvme` is a single NVMe for the VM disks, built for speed rather than redundancy. `zpool_backup` is an HDD for backups.

The dataset properties are set per role: `recordsize: 16k` for the random access of VMs and volumes, `recordsize: 128k` for sequential backups, `compression: lz4` everywhere, `atime: off` against unnecessary writes.

Alongside that, the ARC is capped so ZFS does not take RAM away from guest memory, and ZED sends an email on pool events.

```text
zfs_arc_max=17179869184   # 16 GiB
zfs_arc_min=1073741824    #  1 GiB
ZED_EMAIL_ADDR=admin@example.com
```

## Proxmox storage on ZFS

Finally Proxmox receives the pools as storage. Each mountpoint is registered as a `dir` storage with the matching content type.

```yaml
pve_storages:
  - name: vms
    type: dir
    path: /mnt/zpool_nvme
    content: ["images", "rootdir", "vztmpl", "iso", "snippets"]
  - name: k8s
    type: dir
    path: /mnt/zpool_k8s
    content: ["images", "rootdir"]
  - name: backup
    type: dir
    path: /mnt/zpool_backup
    content: ["backup"]
```

After that the Proxmox UI knows three storages, each on its ZFS pool. VM disks land on the NVMe, Kubernetes volumes on the mirrored SSD pool, backups on the HDD.

## Reproducible flow

```text
bare metal
  -> preseeded Debian (vmbr0, SSH key)
    -> Proxmox VE (lae.proxmox, IOMMU, ZED)
      -> ZFS pools (mrlesmithjr.zfs, by-id, tuning)
        -> Proxmox storage (dir on /mnt/zpool_*)
          -> VM templates and VMs
```

Each stage is a playbook that can be re-run safely. The external roles are version-pinned, yamllint and ansible-lint run in GitHub Actions and as a pre-commit hook.

## Summary

- The whole host is built declaratively with Ansible, every step idempotent
- A preseeded Debian image remasters the netinst ISO via `cpio` and `xorriso`
- Partitioning stays interactive as a guard against accidental overwriting
- The post-install script creates the `vmbr0` bridge and installs the SSH key
- `lae.proxmox` sets up Proxmox, including IOMMU and `vfio` binding for GPU passthrough
- ZFS pools use stable `by-id` devices instead of `sdX`
- `recordsize`, `compression` and `atime` are set per role, the ARC is capped
- Proxmox receives the pools as `dir` storage with matching content types
- External roles are version-pinned, linting runs in CI and pre-commit
