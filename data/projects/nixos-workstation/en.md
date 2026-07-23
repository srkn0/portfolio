---
title: "NixOS Workstation Setup"
description: "Declarative NixOS configuration for two laptops. Flakes, Home Manager, host-specific modules, disko disk layouts, Secure Boot/TPM on the XPS 17 and a custom cx TUI for recurring commands."
tags: [nixos, nix, home-manager, flakes, dotfiles, linux]
date: 2026-07-23
repo: https://github.com/srkn0/nixbase
---

## Overview

My workstation configuration has moved from an Ansible bootstrap into a declarative NixOS setup. The goal is a reproducible laptop state: operating system, desktop, shell, terminal, editor, dotfiles and developer tools live in the repository and are applied through a flake.

The setup currently coexists with the older Ansible repository. Ansible remains useful as bootstrap history, while NixOS increasingly owns the system and user environment.

## Stack & Architecture

- NixOS 26.05 with flakes as the system base
- Home Manager for the user environment, dotfiles, shell, terminal and editor
- Two hosts: Dell XPS 17 and ThinkPad X230
- disko for declarative disk layouts
- lanzaboote and sbctl for Secure Boot on the XPS 17
- TPM2 auto-unlock for LUKS on the XPS 17
- GNOME with PaperWM, PipeWire, Docker and Mullvad as shared system modules
- mise for language runtimes and CLI tools outside Nix
- GitHub Actions for `nix flake check` and host builds

**Host structure:** Each laptop has its own directory under `hosts/` with `default.nix`, `hardware.nix` and `disko.nix`. Shared system building blocks live under `modules/system/`, the user environment under `modules/home/`. Host-specific differences stay visible without duplicating the base configuration.

**Home Manager:** Shell, Starship, Atuin, zoxide, direnv, fzf, Kitty, tmux, Neovim and packages are managed through Home Manager. The Neovim configuration lives as a normal config directory in the repo and is deployed through `xdg.configFile`.

**mise intentionally outside Nix:** Runtime versions for Go, Node, Python and Kubernetes tools live in `mise/config.toml`. A Home Manager activation hook runs `mise install` best-effort, skips it while offline and never lets it fail a rebuild.

## cx Command Library

A small custom Go tool called `cx` collects recurring commands in YAML packs. The TUI searches categories and descriptions, creates new commands, edits existing entries and generates shell aliases. The packs live directly in the repository under `config/cx/`, so changes are visible immediately and do not require a Nix rebuild.

## Verification

CI runs `nix flake check --no-build` and builds both host configurations as `nixosConfigurations.<host>.config.system.build.toplevel`. This catches syntax errors, broken modules and non-buildable host profiles before rollout.
