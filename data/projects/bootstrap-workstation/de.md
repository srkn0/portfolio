---
title: "Ansible Bootstrap für meine Workstation"
description: "Älteres, weiterhin koexistierendes Ansible-Bootstrap für eine Linux/WSL2-Workstation. Ansible verwaltet die OS-Ebene, mise alle CLI-Tools und Language-Runtimes; unter WSL wird der Nerd Font auf dem Windows-Host installiert."
tags: [ansible, mise, wsl, dotfiles, neovim, zsh, molecule]
date: 2026-05-30
category: workstation
featured: 6
repo: https://github.com/srkn0/bootstrap-workstation
---

## Überblick

Ein Befehl macht aus einer frischen Maschine mein Arbeits-Setup. Das Repo ist der Vorgänger meines NixOS-Setups und läuft aktuell noch koexistent, vor allem für WSL2- und klassische Linux-Umgebungen. Das Design teilt die Zuständigkeit danach auf, was ein Tool braucht: alles, was Root benötigt oder ein System-Package ist, läuft über Ansible, alles im User-Space und versionsgepinnt über mise. Das Repo reproduziert einen bekannten Maschinen-Zustand, statt Installationen ad hoc zu skripten.

## Credits

Die Docker-Engine wird über die Role [geerlingguy.docker](https://github.com/geerlingguy/ansible-role-docker) installiert. Die Neovim-Konfiguration ist mein [LazyVim](https://www.lazyvim.org)-Setup, unverändert übernommen. Die Shell baut auf [Oh My Zsh](https://github.com/ohmyzsh/ohmyzsh) und [Starship](https://starship.rs) auf.

## Stack & Architektur

- Ansible für die OS-Ebene (eine Role, eine `site.yml`, getaggte Tasks)
- mise für CLI-Tools und Language-Runtimes, gepinnt in `config.toml`
- Homebrew (linuxbrew) als Ad-hoc-Escape-Hatch
- zsh mit Oh My Zsh und Starship
- Neovim (LazyVim) als Konfiguration eingebunden
- Molecule und ein `tests/Dockerfile` zur Verifikation

**Ansible-Ebene:** kümmert sich um apt-Packages, Fonts, den Docker-Daemon, die Installation von mise und das Ausrollen aller Dotfiles. Jeder Block ist per Tag wählbar und über eine `with_*`-Variable schaltbar.

**mise-Ebene:** eine eingecheckte `config.toml` listet jedes Tool und jede Runtime (Go, Node mit pnpm, Python, das Kubernetes-Tooling, Terraform) auf eine exakte Version fest. `mise install` erledigt den Rest.

**Idempotenz:** ein zweiter Lauf meldet keine Änderungen. Konfigurations-Bäume werden per `copy` ausgerollt, Tool-Installationen über einen Missing-Tool-Check abgesichert.

## Terminal und Font, pro OS

Die Role erkennt WSL zur Laufzeit und verzweigt, damit der Font auf beiden Systemen im Terminal landet. Unter nativem Linux installiert sie den Font nach `~/.local/share/fonts`, installiert das Terminal kitty und schreibt eine `kitty.conf` mit gesetztem Font. Unter WSL wird der Font stattdessen auf dem Windows-Host installiert: die `.ttf`-Dateien werden ins Windows-Font-Verzeichnis des Users kopiert, unter `HKCU` registriert, und die laufende Windows-Terminal-`settings.json` wird in place gepatcht, nur Font-Face und -Size, mit Timestamp-Backup. Andere Terminal-Einstellungen bleiben unangetastet. Font-Name und -Size sind geteilte Variablen, eine Änderung gilt also für beide Pfade.

## Versionen und Updates

Jedes Tool ist auf eine exakte Version gepinnt. Renovate hält sie aktuell: der mise-Manager, die Manager für GitHub Actions und pre-commit, plus ein Custom-Regex-Manager für das eine Tool, das mise nicht installieren kann. Release Notes hängen an jedem PR; Minor- und Patch-Bumps mergen automatisch, Major-Bumps von Runtimes bleiben zur Prüfung offen.

## broot

broot ist das einzige Tool, das nicht über mise läuft. Sein Release ist ein einzelnes Zip, das pro Plattform eine Binary in getrennten Unterordnern bündelt; mise's Backends können das nicht auflösen und griffen die falsche. Ein eigener Task lädt das Zip und extrahiert die zur Architektur passende Linux-Binary nach `~/.local/bin`.
