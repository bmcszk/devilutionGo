---
# devilGo-76ei
title: 'MPQ_FILE_IMPLODE (PKWARE DCL) records are undecodable: retail DIABDAT/hellfire payloads return verbatim'
status: completed
type: bug
priority: high
created_at: 2026-11-03T00:50:39Z
updated_at: 2026-11-03T09:15:54Z
---

Discovered during the asset verification wave (devilGo-5wsb, findings F4/F5).

## Evidence

Block/hash tables decrypted with the StormLib table keys give this record census:

| archive | records | IMPLODE 0x100 | COMPRESS 0x200 | stored |
|---|---|---|---|---|
| DIABDAT.MPQ | 2910 | 1752 | 0 | 1158 |
| hellfire.mpq | 584 | 582 | 0 | 2 |
| devilutionx.mpq | 210 | 0 | 210 | 0 |
| Patch_rt.mpq | 16 | 15 | 1 | 0 |

go-mpq v0.3.1 decodes only MPQ_FILE_COMPRESS payloads (go-mpq@v0.3.1/mpq.go:518-563); an IMPLODE payload is returned verbatim with nil error. Extracting ui_art\mainmenu.pcx from DIABDAT.MPQ yields 9161 bytes starting 0x39 0x69 0xEF 0xFA; assets.DecodePCX rejects it (invalid manufacturer 0x39, expected 0x0A). Upstream routes the flag through pk_explode_sector (devilutionX build-v2/_deps/mpqfs-src/src/mpq_stream.c:153-160).

## Why it blocks the wave

- devilGo-5wsb AC1 (100% extraction smoke) and AC4 (ui_art PCX pixel-exact) cannot pass: retail bytes are not recoverable.
- Every retail sprite/PCX/WAV in DIABDAT/hellfire is affected (1752 + 582 records).

## Fix shape (decide at triage)

1. Patch/vendor go-mpq to route MPQ_FILE_IMPLODE through its existing PKWARE explode (go-mpq@v0.3.1/pkware.go) — smallest change, keeps pure Go; or
2. Carry MPQ reading in the build-time extractor we own (docs/go-libraries.md already lists PKWARE DCL as a must-port-from-C++ gap); or
3. Fork the reader wholesale.

Also needs: a fallback asset-name listfile (retail archives ship none) so extraction can be driven by name at all.

## Proof of Work

Fingerprint: `asset-verification-wave slice3 fingerprint: pkware dcl explode enable AC1`
Commits: `c238493` (DCL decoder), `ccfd719` (read path + AC1/AC4 proof + PoW doc).

### Root cause (confirmed, not inferred)

`go-mpq@v0.3.1/crypt.go:89-117`: `decryptBytes` copies a non-4-aligned buffer,
pads the *copy*, and never writes it back, so `decryptAndDecompressSectors`
decrypts nothing for any sector whose length is not a multiple of four; and
`mpq.go:518-563` dispatches only `MPQ_FILE_COMPRESS`, so `MPQ_FILE_IMPLODE`
payloads returned verbatim. That is the `39 69 EF FA` and the
`zlib: invalid header` this bean reported.

### Fix taken (option 2 — own the read path, no fork)

`internal/assets/dcl.go` — `ExplodeDCL(src, outSize)`: PKWARE DCL explode
(binary model only; every record in the four archives declares model 0), tables
ported verbatim from `mpq_explode.h:56-108`, LSB-first bit walk, overlapping
byte-wise match copy, output truncated to the recorded size.

`internal/assets/mpqarchive.go` — read-only MPQ reader: header, hash/block
table decryption, name lookup, sector offset table (`key-1`) and per-sector
`key+i` decryption on complete words only, zlib/bzip2/PKWARE dispatch, stored
records trimmed to the recorded size. go-mpq remains only the writer for the
synthetic archives the unit tests build (`writeMPQ`); no fork, no vendor, no cgo.

Each rule is pinned by a retail record that only passes when the rule is right:

| rule | source | record that proves it |
|---|---|---|
| decrypt only when `MPQ_FILE_ENCRYPTED` | `SBaseCommon.cpp:490-530` | stored DIABDAT records |
| `key-1` table, `key+i` sectors, complete words | mpqfs `mpq_stream.c:341-402` | `ui_art\mainmenu.pcx`, 1265-byte sector |
| sector only when `(IMPLODE\|COMPRESS) && !SINGLE_UNIT` | `mpq_stream.c:341` | hellfire `gendata\*.smk`, flags `80000000` |
| stored when payload >= output | `mpq_stream.c:96-99` | same `.smk`: 292/420 bytes past their size |

### RED

`dcl_test.go` was written first and failed against a stub returning the input;
both vectors are real decrypted sectors quoted as hex:

| vector | stream | plaintext |
|---|---|---|
| `Patch_rt.mpq` `(listfile)` | 156 B, dict 1024 | 317 B readable text, 15 names CRLF |
| `DIABDAT.MPQ` `monsters\darkmage\dmagew.cl2` | 32 B, dict 4096 | 96 B, `20 00 00 00 27 00 00 00 …` |

The archive-level tests were also RED before the seam fix: `ReadFile` returned
the derived-key XOR of stored payloads (`TestMPQStackPriorityOverride` expected
`high-spawn`, got garbage), because the file key was passed unconditionally
instead of only for encrypted records.

### GREEN — AC1 (100% decode, four archives)

```
DG_ASSET_ARCHIVES="/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire" \
  go test ./internal/assets/ -run TestRealArchivesDecodeEveryRecord -v -count=1
DIABDAT.MPQ:    2911 record(s) (stored 1158, imploded 1752, compressed 0,   unused 1), 2910 decoded, 0 failure(s)
Patch_rt.mpq:     16 record(s) (stored 0,    imploded 15,   compressed 1,   unused 0),   16 decoded, 0 failure(s)
devilutionx.mpq: 210 record(s) (stored 0,    imploded 0,    compressed 210, unused 0),  210 decoded, 0 failure(s)
hellfire.mpq:    584 record(s) (stored 2,    imploded 582,  compressed 0,   unused 0),  584 decoded, 0 failure(s)
--- PASS: TestRealArchivesDecodeEveryRecord (4.90s)
```

3720 addressable records, 3720 decoded to their recorded size, 0 failures. A
used record that no name resolves is a failure, not a skip, so a truncated name
inventory cannot pass silently. The single `Flags == 0` DIABDAT block has no
hash entry and is skipped as unused.

The name gap this bean also raised is closed: `(listfile)` is absent from
DIABDAT/hellfire (encryption keys on the base file name, so a nameless record
cannot even be decrypted). `internal/testdata/assets/mpq/{diabdat,hellfire}.txt`
hold the resolved names as text (2910 + 584), resolved by `HashA`/`HashB`
equality against the zezula.net listfiles; provenance and sha256s in that
directory's `README.md`.

### AC4 blocker cleared

```
go test ./internal/assets/ -run TestRealMainMenuPCXDecodes -v -count=1
--- PASS: TestRealMainMenuPCXDecodes (0.00s)
```
21788 bytes in, manufacturer `0x0A`, `DecodePCX` -> 640x480, 307200 pixels,
palette present. Pixel-exactness against a C++ dump is still devilGo-5wsb AC4's
own criterion and is not claimed here.

### Gates

```
go build ./...                          rc 0
gofmt -l internal/ cmd/                 (empty)
go vet ./...                            rc 0
scripts/lint.sh ./internal/assets/...   0 issues
go test ./... -count=1                  all ok
```

### Environment amendment

`spawn.mpq` is absent in this environment (no shareware install), so AC1's
"spawn.mpq" leg is covered by the environment-absent amendment recorded in
`docs/pow/asset-verification-wave.md`; the walk opens every archive
`DG_ASSET_ARCHIVES` holds and requires 100% decode, so adding a shareware
install needs no code change.
