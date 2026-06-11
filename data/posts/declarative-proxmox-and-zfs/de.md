---
title: "Proxmox und ZFS deklarativ provisionieren"
description: "Bare-Metal-Proxmox per Ansible: ein preseedetes Debian-Image als Installer, Proxmox VE mit IOMMU-GPU-Passthrough, ZFS-Pools über stabile by-id-Geräte, ARC-Tuning und ZED-Alerting, Proxmox-Storage auf ZFS, ein reproduzierbarer Ablauf von der nackten Maschine bis zum VM-Storage."
tags: [proxmox, zfs, ansible, debian, homelab]
date: 2025-02-10
---

## Überblick

Ein Proxmox-Host lässt sich vollständig deklarativ aufbauen, von der nackten Maschine bis zum nutzbaren VM-Storage. Jeder Schritt ist ein Ansible-Playbook oder eine Rolle, jeder Schritt ist idempotent.

Der Ablauf hat vier Stufen: ein preseedetes Debian-Image installiert die Basis, Ansible richtet Proxmox VE ein, ZFS-Pools entstehen über stabile Gerätenamen, und Proxmox bekommt diese Pools als Storage. Das vollständige Setup liegt im Projekt [homelab-proxmox](/projects/homelab-proxmox).

## Debian-Preseed-Image

Der Bare-Metal-Einstieg ist ein preseedetes Debian-Netinst-Image. Eine Ansible-Rolle remastert das offizielle ISO: Sie rendert eine `preseed.cfg` und ein Post-Install-Skript und packt beides per `cpio` in die Initrd, anschließend wird das ISO mit `xorriso` neu geschrieben.

```bash
ansible-playbook playbooks/build_preseed_iso.yml \
  -e input_iso=debian-13-amd64-netinst.iso \
  -e output_iso=preseed-debian-13-amd64-netinst.iso
```

Die `preseed.cfg` beantwortet Locale, Zeitzone, NTP, Paketauswahl und den SSH-Server. Die Partitionierung bleibt bewusst interaktiv, damit das Zielmedium manuell bestätigt wird und keine Platte versehentlich überschrieben wird.

Das Post-Install-Skript bringt den Host in den Zustand, den Proxmox erwartet. Es ersetzt `systemd-timesyncd` durch `chrony`, legt die `vmbr0`-Bridge an und hinterlegt den SSH-Key für den Root-Zugang.

```text
auto vmbr0
iface vmbr0 inet static
    address 192.168.1.10/24
    gateway 192.168.1.1
    bridge-ports <nic>
    bridge-stp off
    bridge-fd 0
```

Root-Passwort und `authorized_keys` kommen aus Variablen. Ergebnis ist ein minimaler Debian-Host mit fertiger Bridge, erreichbar per SSH-Key, bereit für Proxmox.

## Proxmox VE

Proxmox wird über die Rolle `lae.proxmox` eingerichtet, in `requirements.yml` auf eine Version gepinnt und per ansible-galaxy installiert. Die Rolle setzt das No-Subscription-Repository, entfernt den Subscription-Hinweis und installiert `proxmox-ve`.

Für GPU-Passthrough aktiviert dieselbe Rolle IOMMU und bindet die Karte an `vfio-pci`. Im Homelab ist das eine RTX 3060, die später in eine Workload-VM durchgereicht werden soll.

```yaml
pve_pcie_passthrough_enabled: true
pve_iommu_passthrough_mode: true
pve_pci_device_ids:
  - id: "<gpu-pci-id>"
pve_vfio_blacklist_drivers:
  - nouveau
  - nvidia
```

Das setzt `intel_iommu=on iommu=pt` im GRUB, blacklistet die Host-Treiber und lädt die `vfio`-Module. Die Karte gehört damit dem Hypervisor und kann exklusiv an eine VM gebunden werden.

## ZFS-Pools

Die Pools entstehen über die Rolle `mrlesmithjr.zfs`, gesteuert durch eine Liste im Playbook. Geräte werden über ihren `/dev/disk/by-id/`-Namen referenziert, nicht über `sdX`. Der by-id-Name bleibt über Reboots und Steckplatzwechsel stabil; ein Pool aus `sdb` und `sdc` wäre nach dem nächsten Boot ein Glücksspiel.

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

Ein Pool-Layout, drei Rollen. `zpool_k8s` sind SATA-SSDs als Mirror-Vdevs, gestriped, für Kubernetes-Volumes. `zpool_nvme` ist eine einzelne NVMe für die VM-Disks, auf Tempo statt Redundanz ausgelegt. `zpool_backup` ist eine HDD für Backups.

Die Dataset-Eigenschaften sind pro Rolle gesetzt: `recordsize: 16k` für die zufälligen Zugriffe von VMs und Volumes, `recordsize: 128k` für sequentielle Backups, `compression: lz4` überall, `atime: off` gegen unnötige Schreibzugriffe.

Daneben wird der ARC begrenzt, damit ZFS dem Gast-Speicher nicht den RAM wegnimmt, und ZED verschickt eine E-Mail bei Pool-Ereignissen.

```text
zfs_arc_max=17179869184   # 16 GiB
zfs_arc_min=1073741824    #  1 GiB
ZED_EMAIL_ADDR=admin@example.com
```

## Proxmox-Storage auf ZFS

Zum Schluss bekommt Proxmox die Pools als Storage. Jeder Mountpoint wird als `dir`-Storage mit passendem Content-Typ registriert.

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

Danach kennt die Proxmox-Oberfläche drei Storages, jeder auf seinem ZFS-Pool. VM-Disks landen auf der NVMe, Kubernetes-Volumes auf dem gespiegelten SSD-Pool, Backups auf der HDD.

## Reproduzierbarer Ablauf

```text
Bare Metal
  -> preseedetes Debian (vmbr0, SSH-Key)
    -> Proxmox VE (lae.proxmox, IOMMU, ZED)
      -> ZFS-Pools (mrlesmithjr.zfs, by-id, Tuning)
        -> Proxmox-Storage (dir auf /mnt/zpool_*)
          -> VM-Templates und VMs
```

Jede Stufe ist ein Playbook, das sich gefahrlos erneut ausführen lässt. Die externen Rollen sind versions-gepinnt, yamllint und ansible-lint laufen in GitHub Actions und als pre-commit-Hook.

## Zusammenfassung

- Der gesamte Host wird per Ansible deklarativ aufgebaut, jeder Schritt idempotent
- Ein preseedetes Debian-Image remastert das Netinst-ISO per `cpio` und `xorriso`
- Die Partitionierung bleibt interaktiv als Schutz vor versehentlichem Überschreiben
- Das Post-Install-Skript legt die `vmbr0`-Bridge an und hinterlegt den SSH-Key
- `lae.proxmox` richtet Proxmox ein, inklusive IOMMU und `vfio`-Bindung für GPU-Passthrough
- ZFS-Pools nutzen stabile `by-id`-Geräte statt `sdX`
- `recordsize`, `compression` und `atime` werden pro Rolle gesetzt, der ARC wird begrenzt
- Proxmox bekommt die Pools als `dir`-Storage mit passenden Content-Typen
- Externe Rollen sind versions-gepinnt, Linting läuft in CI und pre-commit
