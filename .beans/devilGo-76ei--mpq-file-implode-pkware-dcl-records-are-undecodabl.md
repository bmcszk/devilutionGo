---
# devilGo-76ei
title: 'MPQ_FILE_IMPLODE (PKWARE DCL) records are undecodable: retail DIABDAT/hellfire payloads return verbatim'
status: draft
type: bug
priority: high
created_at: 2026-11-03T00:50:39Z
updated_at: 2026-11-03T00:50:39Z
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
