---
title: "Automatisierte Datenbank-Dumps nach S3"
description: "Tägliche PostgreSQL-Dumps aus Kubernetes per CronJob, Rolling-Retention mit dem MinIO-Client, S3 Object Lock und der Konflikt mit Rotation, hybrider Ansatz mit Archiv, restriktive Bucket Policy."
tags: [kubernetes, postgres, backup, s3, cronjob]
date: 2025-05-20
---

## Ziel

Tägliche, versionierte Dumps aus einem Kubernetes-Cluster nach S3-kompatiblem Speicher. Erzeugung per CronJob mit `pg_dump` und dem MinIO-Client (`mc`).

Als Beispiel dient eine PostgreSQL-Datenbank. Unveränderbarkeit und Zugriffsschutz werden mitberücksichtigt.

## Retention-Schema

Es wird ein Rolling-Prinzip mit exakt `N+1` Dateien verwendet. Bei `N=7` ergeben sich acht Dateien je Datenbank.

- `latest.sql.gz` als aktuelles Backup
- `1.sql.gz` als Backup des Vortags
- `2.sql.gz` als Backup von vorgestern
- bis `7.sql.gz` als ältestes Backup

Die Datei `7.sql.gz` wird beim nächsten Lauf gelöscht, sobald die maximale Retention erreicht ist.

## Ablauf der Rotation

Die Rotation läuft nur, wenn `latest.sql.gz` bereits existiert. Folgende Tabelle zeigt das Verhalten über mehrere Läufe.

| Lauf | Aktionen |
|---|---|
| 1 | Neues Backup als `latest.sql.gz` |
| 2 | `latest -> 1`, neues Backup als `latest.sql.gz` |
| 3 | `1 -> 2`, `latest -> 1`, neues Backup als `latest.sql.gz` |
| 4 | `2 -> 3`, `1 -> 2`, `latest -> 1`, neues Backup als `latest.sql.gz` |

Ab `7.sql.gz` wird das jeweils älteste Backup vor der Rotation entfernt.

### Algorithmus

Der Ablauf je Lauf folgt einer festen Reihenfolge.

- Dump lokal erzeugen, komprimiert als `.sql.gz` in einem temporären Pfad
- Existenz von `latest.sql.gz` im Bucket prüfen
- Fall A: `latest.sql.gz` fehlt, neuer Dump wird direkt als `latest.sql.gz` hochgeladen
- Fall B: `latest.sql.gz` existiert, Rotation wird durchgeführt

Die Rotation in Fall B verläuft so:

- Falls `N.sql.gz` vorhanden, dieses löschen
- Alle vorhandenen Backups um eine Version hochrücken, von `N-1 -> N` bis `1 -> 2`
- `latest.sql.gz -> 1.sql.gz`
- Neuen Dump als `latest.sql.gz` hochladen

Eigenschaften des Verfahrens:

- Stabile, nachvollziehbare Dateistruktur
- Kein versehentliches Überschreiben aktiver Backups
- Maximal `N+1` Dateien je Datenbank, planbare Speicherlast
- Rotation nur bei bereits vorhandenem `latest.sql.gz`

## PoC-Skript

Das folgende Skript erzeugt einen PostgreSQL-Dump und führt die Rotation im Bucket aus. Konfiguration über Umgebungsvariablen.

```bash
#!/bin/bash
set -euo pipefail

# Konfiguration
TMPFILE="/tmp/backup.sql.gz"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
DB_NAME="${DB_NAME:?DB_NAME erforderlich}"
BUCKET="${BUCKET_NAME:?BUCKET_NAME erforderlich}"
S3_PATH="$BUCKET/backups/$DB_NAME"
RETENTION="${RETENTION_DAYS:-7}"
MC_ALIAS="${S3_ALIAS:-s3}"

# MinIO-Client konfigurieren
mc alias set "$MC_ALIAS" "$S3_ENDPOINT" "$S3_ACCESS_KEY" "$S3_SECRET_KEY" >/dev/null

# Dump erzeugen
pg_dump -h "$PGHOST" -U "$PGUSER" "$DB_NAME" | gzip > "$TMPFILE"

# Prüfen, ob latest existiert
if mc stat "$MC_ALIAS/$S3_PATH/latest.sql.gz" >/dev/null 2>&1; then

  # Ältestes Backup entfernen, falls Retention erreicht
  if mc stat "$MC_ALIAS/$S3_PATH/$RETENTION.sql.gz" >/dev/null 2>&1; then
    mc rm "$MC_ALIAS/$S3_PATH/$RETENTION.sql.gz" --quiet
  fi

  # Bestehende Backups um eine Version hochrücken
  for ((i=RETENTION-1; i>=1; i--)); do
    NEXT=$((i+1))
    if mc stat "$MC_ALIAS/$S3_PATH/$i.sql.gz" >/dev/null 2>&1; then
      mc mv "$MC_ALIAS/$S3_PATH/$i.sql.gz" "$MC_ALIAS/$S3_PATH/$NEXT.sql.gz" --quiet
    fi
  done

  # latest auf 1 verschieben
  mc mv "$MC_ALIAS/$S3_PATH/latest.sql.gz" "$MC_ALIAS/$S3_PATH/1.sql.gz" --quiet
fi

# Neuen Dump als latest hochladen
mc cp "$TMPFILE" "$MC_ALIAS/$S3_PATH/latest.sql.gz" --quiet
```

Sensible Werte wie `S3_ACCESS_KEY`, `S3_SECRET_KEY` und `PGPASSWORD` stammen aus einem Secret. `pg_dump` liest das Passwort aus der Umgebungsvariable `PGPASSWORD`. Für andere Datenbanken wird `pg_dump` durch `mysqldump` oder `mongodump` ersetzt, der restliche Ablauf bleibt gleich.

## Unveränderbarkeit und Zugriffsschutz

Backups dürfen nur durch den autorisierten CronJob verändert oder gelöscht werden. Dafür wird der Speicher gegen unbeabsichtigte und unbefugte Schreibzugriffe abgesichert.

### Object Lock in Ceph RGW und MinIO

Ceph RGW und MinIO unterstützen S3 Object Lock einschließlich der Modi GOVERNANCE und COMPLIANCE. Objekte lassen sich damit für einen Zeitraum oder dauerhaft gegen Löschung und Überschreibung schützen.

| Eigenschaft | AWS S3 | Ceph RGW / MinIO |
|---|---|---|
| S3 Object Lock (GOVERNANCE) | Ja | Eingeschränkt |
| S3 Versioning | Ja | Ja |
| Write-Once (WORM) | Ja | Eingeschränkt |
| Bucket Policies (readonly) | Ja | Ja |

Object Lock in Ceph RGW und MinIO bietet nicht die vollen Garantien von AWS S3. Ein Admin-Account kann Objekte weiterhin löschen. Für Schutz vor versehentlichem Löschen reicht das aus, für strikte Compliance nicht ohne Zusatzmaßnahmen.

### Aktivierung beim Bucket-Erstellen

Object Lock muss beim Erstellen des Buckets aktiviert werden. Eine nachträgliche Aktivierung schlägt fehl.

| HTTP-Status | Status Code | Beschreibung |
|---|---|---|
| `400` | MalformedXML | Das XML ist nicht wohlgeformt |
| `409` | InvalidBucketState | Object Lock am Bucket ist nicht aktiviert |

Quelle: [RadosGW BucketOps](https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#put-bucket-object-lock).

## Konflikt zwischen Object Lock und Rotation

Die Rotation benennt Objekte um. Ein `mc mv` ist ein Kopieren mit anschließendem Löschen des Quellobjekts. Auf einem Bucket mit Object Lock ist das Löschen gesperrt, die Rotation schlägt damit fehl.

Object Lock und das Rolling-Schema schließen sich auf demselben Pfad gegenseitig aus.

## Hybrider Ansatz

Die Lösung trennt veränderliche und unveränderliche Objekte in zwei Pfade.

### Veränderliche Rotation

Unter `backups/<db>/` liegen die rotierenden Dateien.

- `latest.sql.gz`, `1.sql.gz` bis `7.sql.gz`
- ohne Object Lock
- Rotation durch das Skript

### Unveränderliches Archiv

Unter `backups/<db>/archive/` liegen zusätzliche Kopien mit Zeitstempel.

```bash
mc cp "$TMPFILE" \
  "$MC_ALIAS/$S3_PATH/archive/$(date +%F).sql.gz" --quiet
```

- Dateien wie `2025-05-20.sql.gz`, `2025-05-21.sql.gz`
- mit Object Lock im GOVERNANCE-Modus
- Retention passend zum Rolling-Schema, etwa sieben Tage

Die Archivkopien werden nur geschrieben, nie umbenannt. Damit ist Object Lock kompatibel.

## Bucket Policy

Nur ein Service-Account darf in den Bucket schreiben. Alle anderen Prinzipale werden für Schreib- und Löschoperationen verweigert.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam:::user/backup-cronjob"
      },
      "Action": ["s3:*"],
      "Resource": [
        "arn:aws:s3:::<bucket>",
        "arn:aws:s3:::<bucket>/*"
      ]
    },
    {
      "Effect": "Deny",
      "NotPrincipal": {
        "AWS": "arn:aws:iam:::user/backup-cronjob"
      },
      "Action": ["s3:PutObject", "s3:DeleteObject"],
      "Resource": ["arn:aws:s3:::<bucket>/*"]
    }
  ]
}
```

Der CronJob nutzt einen exklusiven Access Key. Die Rotation läuft damit über einen zentralen Kontrollpunkt.

## Logging je Tag

Optional protokolliert das Skript jede Aktion in eine Logdatei je Tag.

```bash
LOGFILE="/var/log/db-backup/backup_$(date +%F).log"
mkdir -p "$(dirname "$LOGFILE")"

log() {
  echo "[$(date +'%F %T')] $*" | tee -a "$LOGFILE"
}
```

Im Skript werden die `echo`-Aufrufe durch `log` ersetzt. Die Logdatei dokumentiert Rotation und Upload je Lauf.

## Zusammenfassung

- Rolling-Retention mit `N+1` Dateien, `latest.sql.gz` plus `1..N.sql.gz`
- Rotation nur, wenn `latest.sql.gz` bereits existiert
- Dump per `pg_dump`, gzip-komprimiert, Upload mit `mc`
- `mc mv` löscht das Quellobjekt, daher unvereinbar mit Object Lock
- Ceph RGW und MinIO unterstützen Object Lock, aber ohne die vollen Garantien von AWS S3
- Object Lock muss beim Bucket-Erstellen aktiviert sein
- Hybrider Ansatz: veränderliche Rotation plus unveränderliches Archiv mit Zeitstempeln
- Bucket Policy beschränkt Schreibzugriff auf einen Service-Account
- Logdatei je Tag dokumentiert jeden Lauf
