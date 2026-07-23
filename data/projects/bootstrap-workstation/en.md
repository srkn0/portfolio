---
title: "Ansible bootstrap for my workstation"
description: "Older, still coexisting Ansible bootstrap for a Linux/WSL2 workstation. Ansible owns the OS layer, mise owns all CLI tools and language runtimes; on WSL it installs the Nerd Font onto the Windows host."
tags: [ansible, mise, wsl, dotfiles, neovim, zsh, molecule]
date: 2026-05-30
repo: https://github.com/srkn0/bootstrap-workstation
---

## Overview

One command turns a fresh machine into my working setup. The repo is the predecessor to my NixOS setup and currently still coexists with it, mainly for WSL2 and classic Linux environments. The design splits ownership by what a tool needs: anything requiring root or that is a system package goes through Ansible, everything user-space and version-pinned goes through mise. The repo reproduces a known machine state rather than scripting ad-hoc installs.

## Credits

The Docker engine is installed via the [geerlingguy.docker](https://github.com/geerlingguy/ansible-role-docker) role. The Neovim config is my [LazyVim](https://www.lazyvim.org) setup, vendored verbatim. The shell builds on [Oh My Zsh](https://github.com/ohmyzsh/ohmyzsh) and [Starship](https://starship.rs).

## Stack & Architecture

- Ansible for the OS layer (one role, one `site.yml`, tagged tasks)
- mise for CLI tools and language runtimes, pinned in `config.toml`
- Homebrew (linuxbrew) as an ad-hoc escape hatch
- zsh with Oh My Zsh and Starship
- Neovim (LazyVim) vendored as config
- Molecule and a `tests/Dockerfile` for verification

**Ansible layer:** handles apt packages, fonts, the Docker daemon, installing mise, and deploying every dotfile. Each block is tag-selectable and toggled by a `with_*` variable.

**mise layer:** one committed `config.toml` lists every tool and runtime (Go, Node with pnpm, Python, the Kubernetes tooling, Terraform) at an exact version. `mise install` does the rest.

**Idempotency:** a second run reports no changes. Config trees are deployed with `copy`, tool installs are guarded by a missing-tool check.

## Terminal and font, per OS

The role detects WSL at runtime and branches so the font is wired into the terminal on both systems. On native Linux it installs the font into `~/.local/share/fonts`, installs the kitty terminal, and writes `kitty.conf` with the font set. On WSL it installs the font onto the Windows host instead: it copies the `.ttf` files into the per-user Windows font directory, registers them under `HKCU`, and patches the live Windows Terminal `settings.json` in place to set only the font face and size, with a timestamped backup. Other terminal settings are left untouched. The font name and size are shared variables, so a change applies to both paths.

## Versions and updates

Every tool is pinned to an exact version. Renovate keeps them current: the mise manager, the GitHub Actions and pre-commit managers, plus a custom regex manager for one tool that mise cannot install. Release notes are attached to each PR; minor and patch bumps auto-merge, runtime majors are held for review.

## broot

broot is the one tool not managed by mise. Its release is a single zip that bundles a binary per platform in separate subfolders, which mise's backends cannot disambiguate; they picked the wrong one. A dedicated task downloads the zip and extracts the architecture-matched Linux binary into `~/.local/bin`.
