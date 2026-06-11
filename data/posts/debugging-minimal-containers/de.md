---
title: "Shell-Tricks fürs Debugging in minimalen Containern"
description: "DNS-Auflösung mit getent, Werkzeugerkennung mit command -v, TCP-Verbindungen über /dev/tcp, Prozess- und Socket-Inspektion über /proc, Resolver-Konfiguration, rekursives Globbing und Ephemeral-Debug-Container mit kubectl debug."
tags: [bash, shell, kubernetes, debugging]
date: 2025-06-03
---

## Ausgangslage

Minimale, distroless oder slim aufgebaute Images enthalten oft keine Debug-Werkzeuge. `dig`, `nslookup`, `curl`, `netcat` und `ping` fehlen häufig. Die folgenden Abschnitte sammeln Shell-eigene Mittel, die ohne Zusatzinstallation auskommen.

## Verfügbare Werkzeuge erkennen

Vor jedem Debugging wird geprüft, welche Befehle überhaupt vorhanden sind.

```bash
command -v dig nslookup curl wget nc getent
```

`command -v` ist ein Shell-Builtin und in jeder POSIX-Shell verfügbar. Es gibt nur die gefundenen Pfade aus und überspringt fehlende Befehle. Der Exit-Code zeigt an, ob alle angefragten Befehle existieren.

## DNS-Auflösung ohne dig

Fehlen `dig` und `nslookup`, löst `getent` einen Hostnamen über die Name Service Switch der C-Bibliothek auf.

```bash
getent hosts service.example.com
```

Die Ausgabe enthält die aufgelöste IP-Adresse und den Namen. Die Auflösung folgt der Konfiguration in `/etc/nsswitch.conf` und `/etc/resolv.conf`, also demselben Pfad wie reguläre Anwendungsaufrufe.

Einschränkung: `getent` gehört zu glibc. In Images auf Basis von musl (Alpine) oder BusyBox fehlt es, ebenso in vollständig distroless Images ohne Shell.

## TCP-Verbindung ohne netcat

Bash stellt mit `/dev/tcp/host/port` eine virtuelle Datei bereit, über die sich eine TCP-Verbindung öffnen lässt.

```bash
(exec 3<>/dev/tcp/service.example.com/443) && echo offen || echo zu
```

Der Aufruf öffnet Dateideskriptor 3 als bidirektionale TCP-Verbindung. Gelingt der Verbindungsaufbau, ist der Port erreichbar. Damit lässt sich ohne `netcat` ein Port-Check durchführen.

Einschränkung: `/dev/tcp` ist eine Bash-Erweiterung. In `sh`, `dash` oder BusyBox-`ash` steht es nicht zur Verfügung.

## Prozess-Umgebung lesen

Das Pseudo-Dateisystem `/proc` legt zu jedem Prozess seine Umgebungsvariablen unter `environ` ab.

```bash
tr '\0' '\n' < /proc/1/environ
```

Die Variablen sind durch Nullbytes getrennt; `tr` ersetzt sie durch Zeilenumbrüche. So lässt sich die Konfiguration des Hauptprozesses eines Containers ohne installierte Werkzeuge prüfen.

Einschränkung: Der Zugriff erfordert passende Rechte. Fremde Prozesse sind nur als root oder bei gleichem UID lesbar.

## Lauschende Sockets ohne ss oder netstat

Offene und lauschende Verbindungen stehen unter `/proc/net/tcp` und `/proc/net/tcp6`.

```bash
cat /proc/net/tcp
```

Spalte `local_address` enthält IP und Port hexadezimal kodiert. Status `0A` markiert einen lauschenden Socket. Die Inode-Spalte verknüpft den Socket mit einem Prozess über `/proc/<pid>/fd`.

Einschränkung: Die hexadezimale Kodierung erfordert manuelle Umrechnung. Ohne `ss` oder `netstat` ist die Ausgabe weniger lesbar.

## Resolver-Konfiguration prüfen

Die für die DNS-Auflösung verwendeten Nameserver und Suchdomänen stehen in `/etc/resolv.conf`.

```bash
cat /etc/resolv.conf
```

Die Datei zeigt `nameserver`-Einträge, `search`-Domänen und `options`. In Kubernetes wird sie vom Kubelet anhand der `dnsPolicy` des Pods erzeugt und verweist auf den Cluster-DNS-Dienst.

## Rekursives Globbing in Bash

`shopt -s globstar` aktiviert `**` als rekursives Muster über Verzeichnisbäume hinweg.

```bash
shopt -s globstar
ls **/*.log
```

Ohne `find` lassen sich damit Dateien über alle Unterverzeichnisse hinweg auflisten. Ein einzelnes `echo **` listet den gesamten Baum als flache Alternative zu `tree`.

Einschränkung: `globstar` gibt es ab Bash 4.0. In älteren Bash-Versionen und in POSIX-`sh` fehlt es.

## Ephemeral-Debug-Container an einen Pod hängen

Fehlt im Container jedes Werkzeug, hängt `kubectl debug` einen Ephemeral Container mit vollwertigem Image an einen laufenden Pod.

```bash
kubectl debug -it pod/app --image=nicolaka/netshoot --target=app
```

`--target` teilt den Prozess-Namespace des Zielcontainers, sodass dessen Prozesse und Netzwerk sichtbar sind. Der Ephemeral Container nutzt ein Image mit allen Werkzeugen, ohne das Original-Image zu verändern.

Einschränkung: Ephemeral Containers sind seit Kubernetes 1.25 GA. `--target` setzt Unterstützung für das Teilen von Prozess-Namespaces durch die Container-Runtime voraus.

## Node-Debugging mit kubectl debug

Für die Untersuchung eines Knotens startet `kubectl debug` einen Pod, der das Dateisystem des Knotens einhängt.

```bash
kubectl debug node/node-1 -it --image=nicolaka/netshoot
```

Der Debug-Pod läuft im Host-Namespace des Knotens; dessen Wurzel-Dateisystem ist unter `/host` eingehängt. Damit lassen sich Logs, Konfiguration und Prozesse des Knotens prüfen, ohne ihn per SSH zu betreten.

Einschränkung: Der Aufruf erfordert Rechte zum Erzeugen privilegierter Pods auf dem Knoten.

## Zusammenfassung

- `command -v` prüft als Builtin, welche Werkzeuge vorhanden sind
- `getent hosts` löst DNS über glibc auf, fehlt aber unter musl und BusyBox
- `/dev/tcp/host/port` ersetzt einen Port-Check, nur in Bash verfügbar
- `/proc/<pid>/environ` zeigt die Umgebung eines Prozesses
- `/proc/net/tcp` listet lauschende Sockets ohne `ss` oder `netstat`
- `/etc/resolv.conf` zeigt Nameserver und Suchdomänen
- `shopt -s globstar` aktiviert rekursives `**` ab Bash 4.0
- `kubectl debug` hängt einen Ephemeral Container an einen Pod, GA seit 1.25
- `kubectl debug node/<name>` untersucht einen Knoten über ein vollwertiges Image
