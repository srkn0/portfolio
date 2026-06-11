---
title: "MinIO: Versionierung, Löschen und Retention"
description: "Löschen versionierter S3-Objekte mit dem MinIO Client, Delete-Marker, Löschvarianten über version-id und versions, Aufräumen nach Alter, Object-Lock-Retention und cross-Provider-Beobachtungen."
tags: [minio, s3, object-storage, versioning]
date: 2025-07-17
---

## Löschen versionierter Objekte

Wird ein versioniertes S3-Objekt mit `mc rm` ohne weitere Parameter gelöscht, erfolgt kein physisches Löschen.

```bash
mc rm <alias>/<bucket>/<object>
```

Stattdessen wird ein logischer Delete-Marker als aktuelle Version gesetzt. Das Objekt gilt in normalen Befehlen als nicht vorhanden, die alten Versionen bleiben erhalten.

<details><summary>mc rm help</summary>

```
USAGE:
  mc rm [FLAGS] TARGET [TARGET ...]

FLAGS:
  --versions                       remove object(s) and all its versions
  --recursive, -r                  remove recursively
  --force                          allow a recursive remove operation
  --dangerous                      allow site-wide removal of objects
  --rewind value                   roll back object(s) to current version at specified time
  --version-id value, --vid value  delete a specific version of an object
  --incomplete, -I                 remove incomplete uploads
  --dry-run                        perform a fake remove operation
  --stdin                          read object names from STDIN
  --older-than value               remove objects older than value in duration string (e.g. 7d10h31s)
  --newer-than value               remove objects newer than value in duration string (e.g. 7d10h31s)
  --bypass                         bypass governance
  --non-current                    remove object(s) versions that are non-current
  --config-dir value, -C value     path to configuration folder (default: "/root/.mc") [$MC_CONFIG_DIR]
  --quiet, -q                      disable progress bar display [$MC_QUIET]
  --disable-pager, --dp            disable mc internal pager and print to raw stdout [$MC_DISABLE_PAGER]
  --no-color                       disable color theme [$MC_NO_COLOR]
  --json                           enable JSON lines formatted output [$MC_JSON]
  --debug                          enable debug output [$MC_DEBUG]
  --resolve value                  resolves HOST[:PORT] to an IP address. Example: minio.local:9000=10.10.75.1 [$MC_RESOLVE]
  --insecure                       disable SSL certificate verification [$MC_INSECURE]
  --limit-upload value             limits uploads to a maximum rate in KiB/s, MiB/s, GiB/s. (default: unlimited) [$MC_LIMIT_UPLOAD]
  --limit-download value           limits downloads to a maximum rate in KiB/s, MiB/s, GiB/s. (default: unlimited) [$MC_LIMIT_DOWNLOAD]
  --custom-header value, -H value  add custom HTTP header to the request. 'key:value' format.
  --help, -h                       show help
```

</details>

## Verhalten bei versionierten Objekten

Beim Löschen ohne Parameter wird ein Delete-Marker gesetzt statt physisch gelöscht. Alte Versionen bleiben über `--versions` oder `--version-id` / `--vid` einsehbar und löschbar. Bei aktivem Object Lock ist das Löschen bestimmter Versionen erst nach Ablauf der Retention möglich.

## Löschvarianten

Die folgenden Varianten decken einzelne Versionen, alle Versionen, Filter nach Alter und automatisiertes Aufräumen ab.

| Variante | Parameter | Beschreibung |
| --- | --- | --- |
| Einzelne Version löschen | `--vid` / `--version-id` | Löscht gezielt eine bestimmte Objektversion. |
| Alle Versionen löschen | `--versions` | Entfernt alle Versionen eines Objekts. |
| Nach Alter löschen | `--versions --older-than <dauer>` | Entfernt alte Versionen nach Zeitfilter. |
| Automatisiertes Löschen | `mc find --exec` | Kombiniert `find` mit `rm` für mehrere Objekte. |

### Einzelne Version löschen

Mit `--version-id` bzw. `--vid` wird gezielt eine bestimmte Version entfernt. Das ist nützlich, um einzelne Versionen nach Ablauf der Retention zu löschen.

```bash
mc rm --version-id "<version-id>" "<alias>/<bucket>/<object>"
```

Ausgabe:

```bash
Removed `<alias>/<bucket>/<object>` (versionId=<version-id>).
```

Versionen, für die die Retention noch gilt, können nicht gelöscht werden.

### Mehrere Versionen löschen

Mit `--versions` werden alle Versionen eines Objekts gelöscht. Kombiniert mit `--older-than` eignet sich dies zum Aufräumen alter Backups oder Dumps.

```bash
mc rm --versions --force --older-than 32d <alias>/<bucket>/psql/psql-hourly-dump.crypt
```

Beispielausgabe:

```bash
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
```

Kontrolle der verbleibenden Versionen:

```bash
mc ls --versions <alias>/<bucket>/psql/psql-hourly-dump.crypt | head
[2025-10-02 15:55:56 UTC] 2.2GiB STANDARD 0yGG7Cg9uMBvEOo8jtWXdsosQ1DGWvM v768 PUT psql-hourly-dump.crypt
[2025-10-02 14:56:17 UTC] 2.2GiB STANDARD im1zIfgdAsCNblJ5N.XEcHV0Vn7AOFd v767 PUT psql-hourly-dump.crypt
```

### Automatisches Aufräumen mit mc find

Mit `mc find` lassen sich versionierte Objekte nach Kriterien wie Alter, Name oder Pfad finden und direkt löschen.

```bash
mc find --versions --older-than 30d <alias>/<bucket>/psql \
  --name "psql-*" \
  --exec "mc rm --force --versions {}"
```

Der Befehl durchsucht alle versionierten Objekte unterhalb des Pfads, filtert Dateien älter als 30 Tage und löscht diese über `mc rm --versions`.

## Verhalten mit Retention

Object Lock erzwingt eine Retention pro Version. Das folgende Beispiel zeigt das Zusammenspiel von Delete-Marker, Retention und `--rewind`.

```bash
sh-5.1# mc rm <alias>/<bucket>/test
Created delete marker `<alias>/<bucket>/test` (versionId=cst-Vp.s5NsPE0w0tDcCqY5iSeur.wm).

sh-5.1# mc find --versions --older-than "1m" --json <alias>/<bucket>
{
 "status": "success",
 "type": "",
 "lastModified": "2025-07-17T11:31:24.867Z",
 "size": 8,
 "key": "<alias>/<bucket>/test",
 "versionId": "fDhzzX7NK0wBdasAfNxYKpR4dahL0gi"
}
{
 "status": "success",
 "type": "",
 "lastModified": "2025-07-17T10:35:18.89Z",
 "size": 9,
 "key": "<alias>/<bucket>/test",
 "versionId": "KIpQLx-OwhDPuyJiwG5tR2OJoOFrLfm"
}

# Eine Version, für die die Retention noch gilt, kann nicht gelöscht werden
sh-5.1# mc rm --vid fDhzzX7NK0wBdasAfNxYKpR4dahL0gi <alias>/<bucket>/test
mc: <ERROR> Failed to remove `<alias>/<bucket>/test`. AccessDenied

# Versionen, die vor Konfiguration der Retention erstellt wurden, können gelöscht werden
sh-5.1# mc rm --vid LG85AzWQfQy3PJUulO0BPjU3H6B0VQT <alias>/<bucket>/test
Removed `<alias>/<bucket>/test` (versionId=LG85AzWQfQy3PJUulO0BPjU3H6B0VQT).

# Ein Objekt mit Delete-Marker als current kann nicht per rewind wiederhergestellt werden
sh-5.1# mc rm --rewind 2025-07-17T11:31:24.867Z <alias>/<bucket>/test
mc: <ERROR> Failed to remove `<alias>/<bucket>/test`. Object does not exist
```

Drei Beobachtungen ergeben sich daraus. Eine Version unter aktiver Retention liefert beim Löschen `AccessDenied`. Versionen, die vor Aktivierung der Retention angelegt wurden, lassen sich weiterhin löschen. Ein `--rewind` kann ein Objekt nicht wiederherstellen, dessen aktuelle Version ein Delete-Marker ist.

Nach erneutem Hochladen desselben Objekts wird es wieder aufgelistet, die alten Versionen bleiben erhalten.

```bash
sh-5.1# mc cp test <alias>/<bucket>/test
sh-5.1# mc ls --versions <alias>/<bucket>/test
[2025-07-17 12:07:09 UTC]     0B Fnh1ZBLm8.rHBgewoJl1if6mtZVOg18 v5 DEL test
[2025-07-17 12:05:09 UTC]     8B STANDARD JaZxuOem5sjKaXwT9FmRLgD9XYkrgJQ v4 PUT test
[2025-07-17 11:46:07 UTC]     0B cst-Vp.s5NsPE0w0tDcCqY5iSeur.wm v3 DEL test
[2025-07-17 11:31:24 UTC]     8B STANDARD fDhzzX7NK0wBdasAfNxYKpR4dahL0gi v2 PUT test
[2025-07-17 10:35:18 UTC]     9B STANDARD KIpQLx-OwhDPuyJiwG5tR2OJoOFrLfm v1 PUT test
```

## Versuchsaufbau

Der folgende Aufbau aktiviert Versionierung und Object Lock mit COMPLIANCE-Retention.

```bash
mc mb --with-lock --with-versioning "<alias>/<bucket>"
mc version enable "<alias>/<bucket>"
mc retention set --default COMPLIANCE 1d "<alias>/<bucket>"
mc cp dummy_1.txt "<alias>/<bucket>/dummy_1.txt"
```

COMPLIANCE und GOVERNANCE sind die beiden Object-Lock-Modi. COMPLIANCE verhindert das Löschen bis zum Retain-Until-Datum auch für privilegierte Identitäten, GOVERNANCE erlaubt einen Bypass mit `--bypass`.

## Verhalten nach Ablauf der Retention

Nach Ablauf der eingestellten Retention-Zeit von einem Tag wird das Löschen einer Version geprüft.

```bash
mc rm --version-id "<version-id>" "<alias>/<bucket>/dummy_1.txt"
Removed `<alias>/<bucket>/dummy_1.txt` (versionId=<version-id>).
```

Das Entfernen einer Version nach Ablauf der Retention funktioniert über alle getesteten S3-Implementierungen identisch.

## Cross-Provider-Beobachtung

Über mehrere getestete S3-Implementierungen verhält sich das Löschen nach Ablauf der Retention gleich. Eine Abweichung betrifft nur die Anzeige.

Bei mindestens einer S3-Implementierung meldet `mc retention info` fälschlicherweise, dass kein Object Locking aktiv ist, obwohl die Sperre auf Objektebene greift. Die Objektmetadaten zeigen weiterhin `X-Amz-Object-Lock-Mode` und `X-Amz-Object-Lock-Retain-Until-Date`, und Löschungen vor Ablauf werden blockiert.

```bash
Name      : dummy_1.txt
Date      : 2025-10-01 14:57:47 CEST
Size      : 47 B
ETag      : 4914f1013686fa60afa631df6c7ae00e
VersionID : GRLsktElTlXvsOgtcWQs6ums.0Otwr4
Type      : file
Metadata  :
  Content-Type                       : text/plain
  X-Amz-Object-Lock-Mode             : COMPLIANCE
  X-Amz-Object-Lock-Retain-Until-Date: 2025-10-02T12:57:47.658182707Z
```

Die fehlerhafte Anzeige tritt auf, wenn der Bucket per `mc mb --with-lock` angelegt wird. Beim Anlegen über die Web-UI des Providers tritt sie nicht auf. Löschungen vor Ablauf des Retain-Until-Datums sind blockiert, nach Ablauf funktionieren sie wie erwartet.

## Zusammenfassung

- `mc rm` ohne Parameter setzt einen Delete-Marker, kein physisches Löschen
- `--version-id` / `--vid` löscht genau eine Version, `--versions` alle Versionen
- `--versions --older-than <dauer>` entfernt Versionen nach Alter
- `mc find --versions --older-than <dauer> <path> --exec` räumt mehrere Objekte auf
- Object Lock blockiert das Löschen einer Version bis zum Retain-Until-Datum
- Versionen vor Aktivierung der Retention bleiben löschbar
- `--rewind` stellt kein Objekt wieder her, dessen current ein Delete-Marker ist
- Auf mindestens einer Implementierung meldet `mc retention info` kein Locking, obwohl es auf Objektebene aktiv ist
