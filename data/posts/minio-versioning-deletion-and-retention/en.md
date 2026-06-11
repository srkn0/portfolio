---
title: "MinIO: versioning, deletion and retention"
description: "Deleting versioned S3 objects with the MinIO client, delete markers, deletion variants via version-id and versions, pruning by age, object lock retention and cross-provider observations."
tags: [minio, s3, object-storage, versioning]
date: 2025-07-17
---

## Deleting versioned objects

When a versioned S3 object is deleted with `mc rm` and no further parameters, no physical deletion happens.

```bash
mc rm <alias>/<bucket>/<object>
```

Instead a logical delete marker is set as the current version. The object counts as absent in normal commands, while the old versions remain.

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

## Behaviour with versioned objects

Deleting without parameters sets a delete marker instead of physically removing the object. Old versions remain visible and deletable via `--versions` or `--version-id` / `--vid`. With object lock active, deleting specific versions is only possible after the retention expires.

## Deletion variants

The following variants cover single versions, all versions, filtering by age and automated cleanup.

| Variant | Parameter | Description |
| --- | --- | --- |
| Delete a single version | `--vid` / `--version-id` | Removes one specific object version. |
| Delete all versions | `--versions` | Removes all versions of an object. |
| Delete by age | `--versions --older-than <duration>` | Removes old versions by time filter. |
| Automated deletion | `mc find --exec` | Combines `find` with `rm` for multiple objects. |

### Delete a single version

With `--version-id` or `--vid` one specific version is removed. This is useful for deleting individual versions after the retention expires.

```bash
mc rm --version-id "<version-id>" "<alias>/<bucket>/<object>"
```

Output:

```bash
Removed `<alias>/<bucket>/<object>` (versionId=<version-id>).
```

Versions still under retention cannot be deleted.

### Delete multiple versions

With `--versions` all versions of an object are deleted. Combined with `--older-than` this suits cleaning up old backups or dumps.

```bash
mc rm --versions --force --older-than 32d <alias>/<bucket>/psql/psql-hourly-dump.crypt
```

Example output:

```bash
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
Removed `<alias>/<bucket>/psql/psql-hourly-dump.crypt` (versionId=...)
```

Checking the remaining versions:

```bash
mc ls --versions <alias>/<bucket>/psql/psql-hourly-dump.crypt | head
[2025-10-02 15:55:56 UTC] 2.2GiB STANDARD 0yGG7Cg9uMBvEOo8jtWXdsosQ1DGWvM v768 PUT psql-hourly-dump.crypt
[2025-10-02 14:56:17 UTC] 2.2GiB STANDARD im1zIfgdAsCNblJ5N.XEcHV0Vn7AOFd v767 PUT psql-hourly-dump.crypt
```

### Automated cleanup with mc find

With `mc find`, versioned objects can be located by criteria such as age, name or path and deleted directly.

```bash
mc find --versions --older-than 30d <alias>/<bucket>/psql \
  --name "psql-*" \
  --exec "mc rm --force --versions {}"
```

The command scans all versioned objects below the path, filters files older than 30 days and deletes them via `mc rm --versions`.

## Behaviour with retention

Object lock enforces a retention per version. The following example shows the interplay of delete marker, retention and `--rewind`.

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

# A version still under retention cannot be deleted
sh-5.1# mc rm --vid fDhzzX7NK0wBdasAfNxYKpR4dahL0gi <alias>/<bucket>/test
mc: <ERROR> Failed to remove `<alias>/<bucket>/test`. AccessDenied

# Versions created before retention was configured can be deleted
sh-5.1# mc rm --vid LG85AzWQfQy3PJUulO0BPjU3H6B0VQT <alias>/<bucket>/test
Removed `<alias>/<bucket>/test` (versionId=LG85AzWQfQy3PJUulO0BPjU3H6B0VQT).

# An object whose current version is a delete marker cannot be restored via rewind
sh-5.1# mc rm --rewind 2025-07-17T11:31:24.867Z <alias>/<bucket>/test
mc: <ERROR> Failed to remove `<alias>/<bucket>/test`. Object does not exist
```

Three observations follow from this. A version under active retention returns `AccessDenied` on deletion. Versions created before retention was activated can still be deleted. A `--rewind` cannot restore an object whose current version is a delete marker.

After re-uploading the same object it is listed again, while the old versions remain.

```bash
sh-5.1# mc cp test <alias>/<bucket>/test
sh-5.1# mc ls --versions <alias>/<bucket>/test
[2025-07-17 12:07:09 UTC]     0B Fnh1ZBLm8.rHBgewoJl1if6mtZVOg18 v5 DEL test
[2025-07-17 12:05:09 UTC]     8B STANDARD JaZxuOem5sjKaXwT9FmRLgD9XYkrgJQ v4 PUT test
[2025-07-17 11:46:07 UTC]     0B cst-Vp.s5NsPE0w0tDcCqY5iSeur.wm v3 DEL test
[2025-07-17 11:31:24 UTC]     8B STANDARD fDhzzX7NK0wBdasAfNxYKpR4dahL0gi v2 PUT test
[2025-07-17 10:35:18 UTC]     9B STANDARD KIpQLx-OwhDPuyJiwG5tR2OJoOFrLfm v1 PUT test
```

## Test setup

The following setup enables versioning and object lock with COMPLIANCE retention.

```bash
mc mb --with-lock --with-versioning "<alias>/<bucket>"
mc version enable "<alias>/<bucket>"
mc retention set --default COMPLIANCE 1d "<alias>/<bucket>"
mc cp dummy_1.txt "<alias>/<bucket>/dummy_1.txt"
```

COMPLIANCE and GOVERNANCE are the two object lock modes. COMPLIANCE prevents deletion until the retain-until date even for privileged identities, GOVERNANCE allows a bypass with `--bypass`.

## Behaviour after the retention expires

After the configured retention of one day expires, deleting a version is checked.

```bash
mc rm --version-id "<version-id>" "<alias>/<bucket>/dummy_1.txt"
Removed `<alias>/<bucket>/dummy_1.txt` (versionId=<version-id>).
```

Removing a version after the retention expires works identically across all tested S3 implementations.

## Cross-provider observation

Across several tested S3 implementations, deletion after the retention expires behaves the same. One deviation concerns only the display.

On at least one S3 implementation, `mc retention info` wrongly reports that no object locking is active, although the lock takes effect at the object level. The object metadata still shows `X-Amz-Object-Lock-Mode` and `X-Amz-Object-Lock-Retain-Until-Date`, and deletions before expiry are blocked.

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

The faulty display appears when the bucket is created via `mc mb --with-lock`. When the bucket is created through the provider's web UI it does not appear. Deletions before the retain-until date are blocked, after expiry they work as expected.

## Summary

- `mc rm` without parameters sets a delete marker, no physical deletion
- `--version-id` / `--vid` deletes exactly one version, `--versions` all versions
- `--versions --older-than <duration>` removes versions by age
- `mc find --versions --older-than <duration> <path> --exec` cleans up multiple objects
- Object lock blocks deletion of a version until the retain-until date
- Versions created before retention was activated remain deletable
- `--rewind` does not restore an object whose current version is a delete marker
- On at least one implementation `mc retention info` reports no locking, although it is active at the object level
