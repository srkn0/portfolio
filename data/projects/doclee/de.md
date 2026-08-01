---
title: "doclee"
description: "AI-assisted Full-Stack-Lab nach dem Doctolib-Prinzip: Patienten- und Arztoberflächen, API, Auth, Kalenderlogik, Protobuf-Verträge und Container-/CI-Pfade in einem Monorepo."
tags: [full-stack, tanstack, go, connectrpc, protobuf, nx, healthcare]
date: 2026-08-01
category: lab
repo: https://github.com/srkn0/doclee
---

## Überblick

doclee ist ein umfangreiches Full-Stack-Lab rund um Terminbuchung, Arztprofile, Patientenflüsse und interne Praxis-Workflows. Der Scope ist eine technische Demo mit synthetischen Daten, nicht ein System für echte Patientenvorgänge.

Der Fokus liegt auf Architektur und Lieferfähigkeit: Monorepo-Struktur, getypte Verträge, mehrere Nutzeroberflächen, Service-Grenzen, Container-Builds und ein Deployment-Pfad Richtung Kubernetes.

## Stack & Architektur

- Monorepo mit Nx
- Frontends mit React, TanStack Router und TypeScript
- API in Go
- Protobuf/ConnectRPC für getypte Verträge
- Auth- und Rollenflüsse für mehrere Nutzergruppen
- Container- und CI-Pfade für reproduzierbare Builds

Das Projekt ist groß und zeigt den kompletten Weg von Protobuf-Verträgen über Go-Services bis zu React-Oberflächen und Container-Artefakten.
