# assets/mpq — retail name inventories for the extraction walk

`DIABDAT.MPQ` and `hellfire.mpq` ship **no** `(listfile)` record: they can only
be opened by name, and MPQ encryption is keyed on the base file name, so a name
inventory is a hard precondition for the exhaustive extraction smoke
(`internal/assets/mpq_smoke_test.go`) rather than a convenience.

## Provenance

Source: <https://www.zezula.net/download/listfiles.zip> (Ladislav Zezula's MPQ
listfile collection — the same author as StormLib), downloaded 2026-11-03.

| file | bytes | sha256 |
|---|---|---|
| `listfiles.zip` (not committed) | 24766750 | `23e2a42b9f88b02852a14a1af381a572cf9a337fbcd0378b52cb2fedab6104ae` |
| `Diablo I.txt` (source) | 191105 | `645563977ad55f42076b57c9ce251998ac5328105c70234c94321b0170f61069` |
| `Hellfire.txt` (source) | 24548 | `08e1067dff7babb6e965ec93ad2fea71d36ebb277971cdeb1145c91e042065d3` |

## What is committed here

Only the names that actually resolve, as text, one per line:

| fixture | lines | archive |
|---|---|---|
| `diabdat.txt` | 2910 | `DIABDAT.MPQ` |
| `hellfire.txt` | 584 | `hellfire.mpq` |

A name resolves when `hashString(name, MPQ_HASH_NAME_A)` and
`hashString(name, MPQ_HASH_NAME_B)` equal a live hash table entry's pair, so a
wrong or truncated name is dropped instead of guessed: resolution is
self-verifying and a false positive would take a 2⁻⁶⁴ hash collision. The
external lists carry 7163 (Diablo I) and 979 (Hellfire) names; the remainder are
names for other archive variants and do not exist in these two files.

Rows are ordered by block index so a diff against a different listfile build is
readable, and every block the hash table references appears exactly once —
`TestRealArchivesDecodeEveryRecord` asserts that rather than trusting this file.
