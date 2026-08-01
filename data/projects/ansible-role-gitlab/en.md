---
title: "Ansible Role: GitLab"
description: "New, early-stage Ansible role for self-hosted GitLab instances. It installs GitLab CE or EE from the official package repository and renders a controlled Omnibus configuration for single-node setups."
tags: [ansible, gitlab, self-hosting, molecule, vagrant, qemu]
date: 2026-08-01
category: wip
repo: https://github.com/srkn0/ansible-role-gitlab
---

## Overview

An Ansible role for self-hosted GitLab instances. The goal is a reproducible single-node installation path for GitLab CE or EE, without hiding the official Omnibus package behind custom wrappers. The role configures the matching GitLab Linux package repository, installs the package and manages `/etc/gitlab/gitlab.rb` through a template.

The project started today and is still clearly WIP. The first version focuses on a solid base: package source, package installation, central configuration variables and test paths for Debian- and Red-Hat-based systems.

## Stack & Architecture

- Ansible Core 2.15+ as the runtime
- Official GitLab package repositories via `packages.gitlab.com`
- GitLab CE or EE, selected through `gitlab_edition`
- Omnibus configuration through a Jinja2 template for `gitlab.rb`
- Debian/Ubuntu and RHEL-compatible distributions as target platforms
- Molecule for fast Docker-based scenarios
- Vagrant with QEMU for heavier Ubuntu and Rocky Linux idempotence tests
- go-task and mise for local development commands

**Repository setup:** The role branches by OS family. Debian-based systems get a Deb822 repository with optional APT credentials, while RPM-based systems get a `yum_repository` with the GitLab GPG keys and optional Packagecloud credentials.

**Installation:** Edition and version are explicit variables. Without a version, the latest package for the selected edition is installed; with a version, the package-manager-specific package name is built. The role deliberately does not automate downgrades.

**Configuration:** The template covers the first important Omnibus areas: external URL, time zone, SSH port, Gitaly storage, email sender, SMTP, backups, built-in NGINX, TLS, Let's Encrypt, LDAP, container registry and additional structured or raw Omnibus configuration.

## Tests

The fast path runs through Molecule with a privileged Systemd container. For more realistic checks there are Vagrant/QEMU VMs for Ubuntu 24.04 and Rocky Linux 9. The Taskfile runs syntax checks, ansible-lint and double playbook runs so idempotence stays visible.

## WIP

Next, the tests need to get stricter: more platforms, better verification of the rendered `gitlab.rb`, clearer upgrade paths and probably more variables for GitLab Omnibus options. Until then, the role is a clean starting point rather than a finished abstraction.
