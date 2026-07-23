---
title: "NixOS Workstation Setup"
description: "Deklarative NixOS-Konfiguration für zwei Laptops. Flakes, Home Manager, host-spezifische Module, disko für Disk-Layouts, Secure Boot/TPM auf dem XPS 17 und eine eigene cx-TUI für wiederkehrende Commands."
tags: [nixos, nix, home-manager, flakes, dotfiles, linux]
date: 2026-07-23
repo: https://github.com/srkn0/nixbase
---

## Überblick

Meine Workstation-Konfiguration ist von einem Ansible-Bootstrap zu einem deklarativen NixOS-Setup gewachsen. Ziel ist ein reproduzierbarer Laptop-Zustand: Betriebssystem, Desktop, Shell, Terminal, Editor, Dotfiles und Entwickler-Tools liegen im Repository und werden über einen Flake ausgerollt.

Das Setup läuft aktuell koexistent mit dem älteren Ansible-Repo. Ansible bleibt als Bootstrap-Historie erhalten, NixOS übernimmt aber zunehmend die System- und User-Umgebung.

## Stack & Architektur

- NixOS 26.05 mit Flakes als System-Basis
- Home Manager für User-Umgebung, Dotfiles, Shell, Terminal und Editor
- Zwei Hosts: Dell XPS 17 und ThinkPad X230
- disko für deklarative Disk-Layouts
- lanzaboote und sbctl für Secure Boot auf dem XPS 17
- TPM2-Auto-Unlock für LUKS auf dem XPS 17
- GNOME mit PaperWM, PipeWire, Docker und Mullvad als gemeinsame System-Module
- mise für Language-Runtimes und CLI-Tools außerhalb von Nix
- GitHub Actions für `nix flake check` und Host-Builds

**Host-Struktur:** Jeder Laptop hat ein eigenes Verzeichnis unter `hosts/` mit `default.nix`, `hardware.nix` und `disko.nix`. Gemeinsame System-Bausteine liegen unter `modules/system/`, die User-Umgebung unter `modules/home/`. Dadurch bleiben Host-Unterschiede sichtbar, ohne die Basiskonfiguration zu duplizieren.

**Home Manager:** Shell, Starship, Atuin, zoxide, direnv, fzf, Kitty, tmux, Neovim und Packages werden über Home Manager verwaltet. Die Neovim-Konfiguration liegt als normales Config-Verzeichnis im Repo und wird per `xdg.configFile` ausgerollt.

**mise bewusst außerhalb von Nix:** Runtime-Versionen für Go, Node, Python und Kubernetes-Tools liegen in `mise/config.toml`. Ein Home-Manager-Activation-Hook führt `mise install` best-effort aus, überspringt den Schritt offline und lässt einen Rebuild nicht daran scheitern.

## cx Command Library

Ein eigenes kleines Go-Tool namens `cx` bündelt wiederkehrende Commands in YAML-Packs. Die TUI durchsucht Kategorien und Beschreibungen, legt neue Commands an, editiert bestehende Einträge und generiert Shell-Aliase. Die Packs liegen direkt im Repo unter `config/cx/`, Änderungen sind dadurch sofort sichtbar und brauchen keinen Nix-Rebuild.

## Verifikation

Die CI führt `nix flake check --no-build` aus und baut beide Host-Konfigurationen als `nixosConfigurations.<host>.config.system.build.toplevel`. So fallen Syntaxfehler, kaputte Module und nicht baubare Host-Profile vor dem Rollout auf.
