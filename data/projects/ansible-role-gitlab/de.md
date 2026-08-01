---
title: "Ansible Role: GitLab"
description: "Neue, noch frühe Ansible-Role für selbst gehostete GitLab-Instanzen. Sie installiert GitLab CE oder EE aus dem offiziellen Package-Repository und rendert eine kontrollierte Omnibus-Konfiguration für Single-Node-Setups."
tags: [ansible, gitlab, self-hosting, molecule, vagrant, qemu]
date: 2026-08-01
category: wip
repo: https://github.com/srkn0/ansible-role-gitlab
---

## Überblick

Eine Ansible-Role für selbst gehostete GitLab-Instanzen. Ziel ist ein reproduzierbarer Single-Node-Installationspfad für GitLab CE oder EE, ohne das offizielle Omnibus-Paket hinter eigenen Wrappern zu verstecken. Die Role richtet das passende GitLab-Linux-Repository ein, installiert das Paket und verwaltet `/etc/gitlab/gitlab.rb` über ein Template.

Das Projekt ist ganz frisch gestartet und noch deutlich WIP. Der erste Stand konzentriert sich auf ein belastbares Fundament: Paketquelle, Paketinstallation, zentrale Konfigurationsvariablen und Testpfade für Debian- und Red-Hat-basierte Systeme.

## Stack & Architektur

- Ansible Core 2.15+ als Laufzeit
- Offizielle GitLab-Paketrepositories über `packages.gitlab.com`
- GitLab CE oder EE, wählbar über `gitlab_edition`
- Omnibus-Konfiguration über ein Jinja2-Template für `gitlab.rb`
- Debian/Ubuntu und RHEL-kompatible Distributionen als Zielplattformen
- Molecule für schnelle Docker-basierte Szenarien
- Vagrant mit QEMU für schwerere Ubuntu- und Rocky-Linux-Idempotenztests
- go-task und mise für lokale Entwicklungscommands

**Repository-Setup:** Die Role verzweigt nach OS-Familie. Debian-basierte Systeme bekommen ein Deb822-Repository inklusive optionaler APT-Credentials, RPM-basierte Systeme ein `yum_repository` mit den GitLab-GPG-Keys und optionalen Packagecloud-Credentials.

**Installation:** Edition und Version bleiben explizite Variablen. Ohne Version wird das aktuelle Paket der gewählten Edition installiert; mit Version wird der package-manager-spezifische Paketname gebaut. Downgrades automatisiert die Role bewusst nicht.

**Konfiguration:** Das Template deckt die ersten wichtigen Omnibus-Bereiche ab: externe URL, Zeitzone, SSH-Port, Gitaly-Storage, E-Mail-Absender, SMTP, Backups, integriertes NGINX, TLS, Let's Encrypt, LDAP, Container Registry sowie zusätzliche strukturierte oder rohe Omnibus-Konfiguration.

## Tests

Der schnelle Pfad läuft über Molecule mit einem privilegierten Systemd-Container. Für realistischere Checks gibt es Vagrant/QEMU-VMs für Ubuntu 24.04 und Rocky Linux 9. Die Taskfile führt Syntaxchecks, ansible-lint und doppelte Playbook-Läufe aus, damit Idempotenz sichtbar bleibt.

## WIP

Als nächstes müssen die Tests härter werden: mehr Plattformen, bessere Verifikation der gerenderten `gitlab.rb`, klarere Upgrade-Pfade und wahrscheinlich noch mehr Variablen für GitLab-Omnibus-Optionen. Erst danach ist die Role mehr als ein sauberer Anfang.
