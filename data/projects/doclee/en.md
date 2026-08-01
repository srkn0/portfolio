---
title: "doclee"
description: "AI-assisted full-stack lab inspired by Doctolib: patient and doctor surfaces, API, auth, scheduling logic, Protobuf contracts, and container/CI paths in one monorepo."
tags: [full-stack, tanstack, go, connectrpc, protobuf, nx, healthcare]
date: 2026-08-01
category: lab
repo: https://github.com/srkn0/doclee
---

## Overview

doclee is a large full-stack lab around appointment booking, doctor profiles, patient flows, and internal practice workflows. The scope is a technical demo with synthetic data, not a system for real patient operations.

The focus is architecture and delivery: monorepo structure, typed contracts, multiple user surfaces, service boundaries, container builds, and a deployment path toward Kubernetes.

## Stack & Architecture

- Monorepo with Nx
- Frontends with React, TanStack Router, and TypeScript
- API in Go
- Protobuf/ConnectRPC for typed contracts
- Auth and role flows for multiple user groups
- Container and CI paths for reproducible builds

The project shows the path from Protobuf contracts through Go services to React surfaces and container artifacts.
