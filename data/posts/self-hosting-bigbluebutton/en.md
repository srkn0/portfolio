---
title: "Self-hosting BigBlueButton"
description: "Installing BigBlueButton 3.0 with bbb-install.sh, system requirements and ports, Greenlight as the frontend, your own TLS certificate behind a reverse proxy, handling an internal CA and the full certificate chain."
tags: [bigbluebutton, self-hosting, ubuntu, webrtc]
date: 2025-11-12
---

## BigBlueButton

BigBlueButton is an open-source web conferencing system for online teaching and meetings. Audio, video and screen sharing run over WebRTC. Installation is handled by the `bbb-install.sh` script, which sets up a complete server on a fresh Ubuntu.

A BigBlueButton installation is demanding on resources and belongs on a dedicated host without other services.

## Requirements

`bbb-install.sh` expects a clean server. Existing web servers or packages cause conflicts.

- Ubuntu 22.04 (64-bit), for BigBlueButton 3.0
- 8 CPU cores with good single-thread performance, 16 GB RAM
- 500 GB of disk for recordings, 50 GB without recording
- an FQDN such as `bbb.example.com` with a public IPv4 and IPv6 address
- TCP ports 80 and 443 reachable
- UDP ports 16384 to 32768 reachable for the media streams

The UDP range is decisive. If it is missing in the firewall, the connection is established but audio and video stay absent.

## Installation

Installation runs through a single call. The version is selected via a string: `jammy-300` stands for BigBlueButton 3.0 on Ubuntu 22.04.

```bash
wget -qO- https://raw.githubusercontent.com/bigbluebutton/bbb-install/v3.0.x-release/bbb-install.sh \
  | bash -s -- -v jammy-300 -s bbb.example.com -e admin@example.com -w -g
```

The main options:

| Flag | Meaning |
|---|---|
| `-v jammy-300` | version: BigBlueButton 3.0 on Ubuntu 22.04 |
| `-s bbb.example.com` | FQDN of the server |
| `-e admin@example.com` | email for the Let's Encrypt certificate |
| `-w` | set up the UFW firewall |
| `-g` | install Greenlight as the frontend |
| `-d` | skip the certificate request, use your own certificates |

With `-e` the script requests a Let's Encrypt certificate automatically. This requires the FQDN to point publicly at the server and port 80 to be reachable.

## Greenlight as the frontend

Without a frontend BigBlueButton offers only an API. Greenlight is the bundled web interface for creating rooms, inviting participants and managing recordings. The `-g` flag installs it as a Docker Compose stack alongside the server.

```bash
bbb-conf --check
```

`bbb-conf --check` validates the installation and reports wrong hostnames, missing ports or certificate problems.

## Your own certificate behind a reverse proxy

In internal environments a reverse proxy often terminates TLS, or the certificate comes from an internal CA rather than Let's Encrypt. The automatic request is then skipped, and the `-d` flag takes the provided certificates.

The files are placed before the installation:

```text
/local/certs/
├── fullchain.pem
└── privkey.pem
```

```bash
wget -qO- https://raw.githubusercontent.com/bigbluebutton/bbb-install/v3.0.x-release/bbb-install.sh \
  | bash -s -- -v jammy-300 -s bbb.example.com -d
```

`fullchain.pem` contains the server certificate and all intermediate certificates up to the CA. `privkey.pem` is the matching private key.

## Internal CA: a full chain instead of disabling TLS

When the certificate comes from an internal CA, rendering uploaded presentations fails with a message like `unable to verify the first certificate`. The backend process loads the presentation over its own HTTPS URL and cannot verify the chain.

The cause is an incomplete or untrusted certificate chain, not an overly strict check. The correct fix is to complete the chain and trust the CA, not to disable verification.

```bash
sudo cp internal-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

`fullchain.pem` must contain the intermediate and root certificates of the internal CA. If Greenlight runs in a container, the CA is mounted into the container as well, so that its process trusts the chain too.

Disabling the TLS check, for example via `NODE_TLS_REJECT_UNAUTHORIZED=0`, fixes the symptom but removes the encryption guarantee and does not belong in production.

## Operations

After configuration changes, `bbb-conf --restart` restarts all components. Recordings are processed after the meeting and only appear in Greenlight once processing is complete.

```bash
bbb-conf --restart
bbb-conf --check
```

## Summary

- `bbb-install.sh` sets up BigBlueButton on a clean Ubuntu 22.04
- BigBlueButton 3.0 is selected through the version `jammy-300`
- The host needs 8 cores, 16 GB RAM, TCP 80/443 and UDP 16384-32768
- A missing UDP range prevents audio and video despite an established connection
- `-g` installs Greenlight as the frontend, `-e` requests a Let's Encrypt certificate
- `-d` uses your own certificates from `/local/certs/`, for example behind a reverse proxy
- With an internal CA, `fullchain.pem` must contain the full chain and the CA must be trusted
- Disabling the TLS check fixes the symptom, not the cause
