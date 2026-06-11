---
title: "BigBlueButton selbst hosten"
description: "BigBlueButton 3.0 mit bbb-install.sh installieren, Systemvoraussetzungen und Ports, Greenlight als Frontend, eigenes TLS-Zertifikat hinter einem Reverse Proxy, Umgang mit einer internen CA und der vollständigen Zertifikatskette."
tags: [bigbluebutton, self-hosting, ubuntu, webrtc]
date: 2025-11-12
---

## BigBlueButton

BigBlueButton ist ein quelloffenes Web-Conferencing-System für Online-Unterricht und Meetings. Audio, Video und Bildschirmfreigabe laufen über WebRTC. Die Installation übernimmt das Skript `bbb-install.sh`, das einen kompletten Server auf einem frischen Ubuntu aufsetzt.

Eine BigBlueButton-Installation ist anspruchsvoll im Ressourcenbedarf und gehört auf einen dedizierten Host ohne andere Dienste.

## Voraussetzungen

`bbb-install.sh` erwartet einen sauberen Server. Vorhandene Webserver oder Pakete führen zu Konflikten.

- Ubuntu 22.04 (64-bit), für BigBlueButton 3.0
- 8 CPU-Kerne mit guter Single-Thread-Leistung, 16 GB RAM
- 500 GB Plattenplatz für Aufzeichnungen, 50 GB ohne Recording
- ein FQDN wie `bbb.example.com` mit öffentlicher IPv4- und IPv6-Adresse
- TCP-Ports 80 und 443 erreichbar
- UDP-Ports 16384 bis 32768 für die Medienströme erreichbar

Der UDP-Bereich ist entscheidend. Fehlt er in der Firewall, kommt die Verbindung zustande, aber Audio und Video bleiben aus.

## Installation

Die Installation läuft über einen Aufruf. Die Version wird über einen String gewählt: `jammy-300` steht für BigBlueButton 3.0 auf Ubuntu 22.04.

```bash
wget -qO- https://raw.githubusercontent.com/bigbluebutton/bbb-install/v3.0.x-release/bbb-install.sh \
  | bash -s -- -v jammy-300 -s bbb.example.com -e admin@example.com -w -g
```

Die wichtigsten Optionen:

| Flag | Bedeutung |
|---|---|
| `-v jammy-300` | Version: BigBlueButton 3.0 auf Ubuntu 22.04 |
| `-s bbb.example.com` | FQDN des Servers |
| `-e admin@example.com` | E-Mail für das Let's-Encrypt-Zertifikat |
| `-w` | UFW-Firewall einrichten |
| `-g` | Greenlight als Frontend installieren |
| `-d` | keine Zertifikatsanforderung, eigene Zertifikate verwenden |

Mit `-e` fordert das Skript automatisch ein Let's-Encrypt-Zertifikat an. Das setzt voraus, dass der FQDN öffentlich auf den Server zeigt und Port 80 erreichbar ist.

## Greenlight als Frontend

Ohne Frontend bietet BigBlueButton nur eine API. Greenlight ist die mitgelieferte Weboberfläche zum Anlegen von Räumen, Einladen von Teilnehmern und Verwalten von Aufzeichnungen. Das Flag `-g` installiert sie als Docker-Compose-Stack neben dem Server.

```bash
bbb-conf --check
```

`bbb-conf --check` prüft die Installation und meldet falsch gesetzte Hostnamen, fehlende Ports oder Zertifikatsprobleme.

## Eigenes Zertifikat hinter einem Reverse Proxy

In internen Umgebungen terminiert oft ein Reverse Proxy das TLS, oder das Zertifikat stammt aus einer eigenen CA statt von Let's Encrypt. Dann entfällt die automatische Anforderung, und das Flag `-d` übernimmt die bereitgestellten Zertifikate.

Die Dateien werden vor der Installation abgelegt:

```text
/local/certs/
├── fullchain.pem
└── privkey.pem
```

```bash
wget -qO- https://raw.githubusercontent.com/bigbluebutton/bbb-install/v3.0.x-release/bbb-install.sh \
  | bash -s -- -v jammy-300 -s bbb.example.com -d
```

`fullchain.pem` enthält das Server-Zertifikat und alle Zwischenzertifikate bis zur CA. `privkey.pem` ist der zugehörige private Schlüssel.

## Interne CA: vollständige Kette statt TLS abschalten

Stammt das Zertifikat aus einer internen CA, scheitert das Rendern hochgeladener Präsentationen mit einer Meldung wie `unable to verify the first certificate`. Der Backend-Prozess lädt die Präsentation über die eigene HTTPS-URL und kann die Kette nicht verifizieren.

Die Ursache ist eine unvollständige oder nicht vertraute Zertifikatskette, nicht eine zu strenge Prüfung. Die korrekte Lösung ist, die Kette vollständig zu machen und die CA als vertrauenswürdig zu hinterlegen, nicht die Verifikation abzuschalten.

```bash
sudo cp internal-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

Die `fullchain.pem` muss die Zwischen- und Root-Zertifikate der internen CA enthalten. Läuft Greenlight im Container, wird die CA zusätzlich in den Container eingebunden, sodass auch dessen Prozess der Kette vertraut.

Das Deaktivieren der TLS-Prüfung, etwa über `NODE_TLS_REJECT_UNAUTHORIZED=0`, behebt das Symptom, hebt aber die Verschlüsselungsgarantie auf und gehört nicht in einen Betrieb.

## Betrieb

Nach Änderungen an der Konfiguration startet `bbb-conf --restart` alle Komponenten neu. Aufzeichnungen werden nach dem Meeting verarbeitet und stehen erst nach Abschluss der Verarbeitung in Greenlight bereit.

```bash
bbb-conf --restart
bbb-conf --check
```

## Zusammenfassung

- `bbb-install.sh` setzt BigBlueButton auf einem sauberen Ubuntu 22.04 auf
- BigBlueButton 3.0 wird über die Version `jammy-300` gewählt
- Der Host braucht 8 Kerne, 16 GB RAM, TCP 80/443 und UDP 16384-32768
- Ein fehlender UDP-Bereich verhindert Audio und Video trotz bestehender Verbindung
- `-g` installiert Greenlight als Frontend, `-e` fordert ein Let's-Encrypt-Zertifikat an
- `-d` nutzt eigene Zertifikate aus `/local/certs/`, etwa hinter einem Reverse Proxy
- Bei einer internen CA muss `fullchain.pem` die vollständige Kette enthalten und die CA vertrauenswürdig sein
- TLS-Prüfung abzuschalten behebt das Symptom, nicht die Ursache
