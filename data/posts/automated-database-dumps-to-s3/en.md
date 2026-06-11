---
title: "Automated database dumps to S3"
description: "Daily PostgreSQL dumps from Kubernetes via CronJob, rolling retention with the MinIO client, S3 Object Lock and the conflict with rotation, hybrid approach with an archive, restrictive bucket policy."
tags: [kubernetes, postgres, backup, s3, cronjob]
date: 2025-05-20
---

## Goal

Daily, versioned dumps from a Kubernetes cluster to S3-compatible storage. Created by a CronJob with `pg_dump` and the MinIO client (`mc`).

The example uses a PostgreSQL database. Immutability and access protection are taken into account.

## Retention scheme

A rolling scheme with exactly `N+1` files is used. With `N=7` this results in eight files per database.

- `latest.sql.gz` as the current backup
- `1.sql.gz` as the previous day's backup
- `2.sql.gz` as the backup from two days ago
- through `7.sql.gz` as the oldest backup

The file `7.sql.gz` is deleted on the next run once the maximum retention is reached.

## Rotation flow

Rotation runs only when `latest.sql.gz` already exists. The following table shows the behavior across several runs.

| Run | Actions |
|---|---|
| 1 | New backup as `latest.sql.gz` |
| 2 | `latest -> 1`, new backup as `latest.sql.gz` |
| 3 | `1 -> 2`, `latest -> 1`, new backup as `latest.sql.gz` |
| 4 | `2 -> 3`, `1 -> 2`, `latest -> 1`, new backup as `latest.sql.gz` |

From `7.sql.gz` on, the oldest backup is removed before the rotation.

### Algorithm

Each run follows a fixed order.

- Create the dump locally, compressed as `.sql.gz` in a temporary path
- Check for the existence of `latest.sql.gz` in the bucket
- Case A: `latest.sql.gz` is missing, the new dump is uploaded directly as `latest.sql.gz`
- Case B: `latest.sql.gz` exists, rotation is performed

The rotation in case B proceeds as follows:

- If `N.sql.gz` is present, delete it
- Move all existing backups up one version, from `N-1 -> N` through `1 -> 2`
- `latest.sql.gz -> 1.sql.gz`
- Upload the new dump as `latest.sql.gz`

Properties of the scheme:

- Stable, traceable file structure
- No accidental overwriting of active backups
- At most `N+1` files per database, predictable storage load
- Rotation only when `latest.sql.gz` already exists

## PoC script

The following script creates a PostgreSQL dump and performs the rotation in the bucket. Configuration through environment variables.

```bash
#!/bin/bash
set -euo pipefail

# Configuration
TMPFILE="/tmp/backup.sql.gz"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
DB_NAME="${DB_NAME:?DB_NAME required}"
BUCKET="${BUCKET_NAME:?BUCKET_NAME required}"
S3_PATH="$BUCKET/backups/$DB_NAME"
RETENTION="${RETENTION_DAYS:-7}"
MC_ALIAS="${S3_ALIAS:-s3}"

# Configure the MinIO client
mc alias set "$MC_ALIAS" "$S3_ENDPOINT" "$S3_ACCESS_KEY" "$S3_SECRET_KEY" >/dev/null

# Create the dump
pg_dump -h "$PGHOST" -U "$PGUSER" "$DB_NAME" | gzip > "$TMPFILE"

# Check whether latest exists
if mc stat "$MC_ALIAS/$S3_PATH/latest.sql.gz" >/dev/null 2>&1; then

  # Remove the oldest backup if retention is reached
  if mc stat "$MC_ALIAS/$S3_PATH/$RETENTION.sql.gz" >/dev/null 2>&1; then
    mc rm "$MC_ALIAS/$S3_PATH/$RETENTION.sql.gz" --quiet
  fi

  # Move existing backups up one version
  for ((i=RETENTION-1; i>=1; i--)); do
    NEXT=$((i+1))
    if mc stat "$MC_ALIAS/$S3_PATH/$i.sql.gz" >/dev/null 2>&1; then
      mc mv "$MC_ALIAS/$S3_PATH/$i.sql.gz" "$MC_ALIAS/$S3_PATH/$NEXT.sql.gz" --quiet
    fi
  done

  # Move latest to 1
  mc mv "$MC_ALIAS/$S3_PATH/latest.sql.gz" "$MC_ALIAS/$S3_PATH/1.sql.gz" --quiet
fi

# Upload the new dump as latest
mc cp "$TMPFILE" "$MC_ALIAS/$S3_PATH/latest.sql.gz" --quiet
```

Sensitive values such as `S3_ACCESS_KEY`, `S3_SECRET_KEY` and `PGPASSWORD` come from a secret. `pg_dump` reads the password from the `PGPASSWORD` environment variable. For other databases `pg_dump` is replaced by `mysqldump` or `mongodump`, the rest of the flow stays the same.

## Immutability and access protection

Backups may only be changed or deleted by the authorized CronJob. The storage is therefore protected against accidental and unauthorized write access.

### Object Lock in Ceph RGW and MinIO

Ceph RGW and MinIO support S3 Object Lock including the GOVERNANCE and COMPLIANCE modes. Objects can thus be protected against deletion and overwriting for a period or permanently.

| Feature | AWS S3 | Ceph RGW / MinIO |
|---|---|---|
| S3 Object Lock (GOVERNANCE) | Yes | Limited |
| S3 Versioning | Yes | Yes |
| Write-Once (WORM) | Yes | Limited |
| Bucket policies (read-only) | Yes | Yes |

Object Lock in Ceph RGW and MinIO does not provide the full guarantees of AWS S3. An admin account can still delete objects. This is sufficient as protection against accidental deletion, but not for strict compliance without additional measures.

### Activation at bucket creation

Object Lock must be enabled when the bucket is created. Enabling it afterwards fails.

| HTTP status | Status code | Description |
|---|---|---|
| `400` | MalformedXML | The XML is not well-formed |
| `409` | InvalidBucketState | Object Lock on the bucket is not enabled |

Source: [RadosGW BucketOps](https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#put-bucket-object-lock).

## Conflict between Object Lock and rotation

Rotation renames objects. An `mc mv` is a copy followed by deletion of the source object. On a bucket with Object Lock deletion is blocked, so the rotation fails.

Object Lock and the rolling scheme are mutually exclusive on the same path.

## Hybrid approach

The solution separates mutable and immutable objects into two paths.

### Mutable rotation

The rotating files live under `backups/<db>/`.

- `latest.sql.gz`, `1.sql.gz` through `7.sql.gz`
- without Object Lock
- rotated by the script

### Immutable archive

Additional timestamped copies live under `backups/<db>/archive/`.

```bash
mc cp "$TMPFILE" \
  "$MC_ALIAS/$S3_PATH/archive/$(date +%F).sql.gz" --quiet
```

- files such as `2025-05-20.sql.gz`, `2025-05-21.sql.gz`
- with Object Lock in GOVERNANCE mode
- retention matching the rolling scheme, around seven days

The archive copies are only written, never renamed. This is compatible with Object Lock.

## Bucket policy

Only one service account may write to the bucket. All other principals are denied write and delete operations.

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

The CronJob uses an exclusive access key. Rotation thus runs through a single control point.

## Per-day logging

Optionally the script logs every action to a per-day log file.

```bash
LOGFILE="/var/log/db-backup/backup_$(date +%F).log"
mkdir -p "$(dirname "$LOGFILE")"

log() {
  echo "[$(date +'%F %T')] $*" | tee -a "$LOGFILE"
}
```

In the script the `echo` calls are replaced by `log`. The log file documents rotation and upload per run.

## Summary

- Rolling retention with `N+1` files, `latest.sql.gz` plus `1..N.sql.gz`
- Rotation only when `latest.sql.gz` already exists
- Dump via `pg_dump`, gzip-compressed, upload with `mc`
- `mc mv` deletes the source object, so it is incompatible with Object Lock
- Ceph RGW and MinIO support Object Lock, but without the full guarantees of AWS S3
- Object Lock must be enabled at bucket creation
- Hybrid approach: mutable rotation plus an immutable archive with timestamps
- Bucket policy restricts write access to a single service account
- A per-day log file documents every run
