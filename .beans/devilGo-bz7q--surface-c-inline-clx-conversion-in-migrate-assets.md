---
# devilGo-bz7q
title: 'Surface C: inline CLX conversion in migrate assets'
status: completed
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T22:12:41Z
parent: devilGo-8csd
blocked_by:
    - devilGo-srb3
---

# Surface C: inline CLX conversion in migrate assets

## Goal
Per ADR-019 (Decision part 1): `migrate assets` conversion. The extraction pass converts extracted `.cel`, `.cl2`, `.pcx` records that carry frame-width inventory data into `.clx` using the existing internal/assets converters; records without width data keep today's copy-as-is behavior (extract.go:249-251). One write = ADR-005 embed parity ("the build embeds exactly the files the tool writes").

## Research (pre-validated)
- Parent review APPROVE (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md, task-dgconv-review-brainer).
- docs/research/dg-conversion-inventory.md: today the conversion is a SECOND program (dgo-assets convert), so what the build embeds does not equal what migrate writes — the split the user ruled out.
- Converters proven in internal/assets: cel_to_clx.go, cl2_to_clx.go, plus PCX handling; the l5fp craft covers the manifest loader/width fields.
- Widths source: the graphics inventory entries carry frames/width/width_source (internal/assets/manifest/manifest.go:91-96) — the inventory is the width oracle for convert decisions.

## Scope
- assets pass: for each convertible record whose inventory entry has width data -> CLX output replaces the raw file in the assets dir; manifest records updated (final format clx + the conversion inputs hashed, keeping idempotency summary stable across a second run).
- No-width records: copy as-is, format stays cel/cl2/pcx.
- dgo-assets convert remains available until Surface D (folded CLI path wins if both exist).
- Tests: synthetic fixtures only (never retail MPQ); round-trip the proven converter tests; idempotency test re-run hash stable.

## Acceptance criteria
- [ ] `migrate assets` on synthetic fixtures writes .clx for width-carrying records; copy-as-is otherwise (integration test with fixture CSV/TSV inputs, no MPQ)
- [ ] Idempotent: second run produces identical assets dir + identical hash summary (test)
- [ ] ADR-005 embed parity: a test asserts the embed source dir contains exactly what migrate wrote
- [x] Full gates green

## Libraries

No new library bindings. Everything stays within the pinned catalog (ADR-009); internal/assets converters already committed.

## Skills for coder
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/beans/SKILL.md

## Dependencies & blockers
blockedBy: Surface A. Blocks Surface D. Parent: epic devilGo-8csd.

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.

## IMPL NOTES

- Code SHA: 58f5b2b (branch dg-surface).
- Conversion lives in `cmd/devilutionGo-migrate/assets_conversion.go`; the
  extraction pass calls it after `render` (`extract.go` `extractRecord`), so a
  converted record's payload/outputName/family/format/frames are replaced before
  the manifest entry is built. Converters in `internal/assets` are untouched.
- Width oracle: `conversionInputs`, keyed by the lowercased forward-slash CLX
  target path. An inventory entry's `widths` (else `width`) is the geometry;
  entries without widths are not convertible. Lookup is case/separator
  normalised because retail record names and inventory paths differ in both.
- PCX: the converter takes a frame count, not widths, so the inventory's
  `frames` is passed and no transparent index is supplied (the inventory model
  has no field for one). Recorded as `conversion.format: pcx`.
- `assets extract|assets` gained `--inventory` (default `assets/inventory.yaml`).
  A missing inventory file converts nothing and prints a stderr notice naming
  the path; a present-but-invalid inventory fails the run (`loadInventory` ->
  `Validate`). Fail-closed only on the file the operator supplied, so the
  pipeline still runs before the requirements list is committed.
- `fileEntry` gained `conversion` (format/widths/frames/sha256 of the inputs);
  the entry's own `sha256` stays the hash of the written `.clx`, so
  `assets verify` and the manifest/digest parity hold.
- Folded path wins: `dgo-assets convert` stays (Surface D deletes it); nothing
  invokes it, and the folded `assets convert`/inline pass is what the tool runs.
- Flag-name asymmetry noted for Surface D: `assets verify|list` still spell the
  inventory as `--manifest` (Surface A kept that name); extract spells it
  `--inventory`.
- AC wording deviation: AC 1 says "fixture CSV/TSV inputs"; the conversion
  fixtures are the committed gfx hex goldens (`internal/testdata/gfx/cel`,
  `cl2`) fed through a synthetic MPQ built in the test, like every other
  migrate test. No retail MPQ, no CSV/TSV involvement in conversion.

## Proof of Work

RED first (conversion disabled with a temporary early return, then reverted):

```
--- FAIL: TestAssetsConvertsLegacyRecordsTheInventoryDescribes (0.00s)
    --- FAIL: .../clx_bytes_match_the_proven_converter_goldens (0.00s)
        assets_conversion_test.go:33: read output spawn/data/talkbutton.clx: no such file or directory
    --- FAIL: .../the_legacy_container_is_not_written (0.00s)
        assets_conversion_test.go:39: spawn/data/talkbutton.cel was written, want only its clx conversion
    --- FAIL: .../the_manifest_records_the_conversion_inputs (0.00s)
        assets_conversion_test.go:44: spawn/data/talkbutton.clx is not in the manifest
    --- FAIL: .../the_report_counts_the_conversions (0.00s)
FAIL	github.com/bmcszk/devilutionGo/cmd/devilutionGo-migrate	0.010s
```

Green gates after the implementation:

```
$ go build ./...
BUILD OK
$ gofmt -l .
(empty above = clean)
$ ./scripts/lint.sh
lint: sim purity gate
simpurity: internal/game internal/diablo clean (226 sources)
0 issues.
$ go vet ./internal/...
VET OK
$ go test ./... -count=1
44 packages ok, 0 FAIL (grep -c '^ok' = 44)
$ python3 scripts/beans-lint.py
beans-lint: OK (170 beans)
$ just sqlc-check
no raw SQL outside sqlc-generated code
```

New tests: `TestAssetsConvertsLegacyRecordsTheInventoryDescribes` (CEL and CL2
bytes equal the committed C++ converter goldens; the legacy container is not
written; manifest records the inputs; report counts), 
`TestAssetsCopiesLegacyRecordsTheInventoryDoesNotDescribe`,
`TestAssetsFailsOnAMalformedInventory`,
`TestAssetsConversionRerunIsIdempotent` (identical dir digests + summary sha256,
second run resumes), `TestAssetsConvertedDirHoldsExactlyTheManifestFiles`
(embed-parity: manifest files + manifest.yaml + journal, no legacy leftovers).
