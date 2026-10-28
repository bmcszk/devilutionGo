---
# devilGo-aysp
title: 'PCX decode: row order, BytesPerLine, transparent colour'
status: completed
type: bug
priority: high
created_at: 2026-10-28T16:35:33Z
updated_at: 2026-10-28T16:54:57Z
parent: devilGo-5wsb
---

## Goal

Make `DecodePCX` byte-faithful to `PcxToClx`: reverse the scanline order, honour
`BytesPerLine`, and support the optional transparent colour.

## Evidence (verified in the parent against the pinned C++)

1. **Row order is wrong.** C++ `Source/utils/pcx_to_clx.cpp:123` states "We process the PCX a
   whole frame at a time because the lines are reversed in CEL", and `:169` reads
   `src = &frameBuffer[(frameHeight - (line + 1)) * width]` — frames are emitted bottom-up.
   Go `internal/assets/pcx.go` `DecodePCX` returns pixels in stored scanline order, so every
   decoded PCX frame is vertically flipped relative to the C++ product.
2. **`BytesPerLine` is parsed but ignored.** Go parses it at `pcx.go:128` and never validates or
   uses it. C++ validates `bytesPerLine < width` as an error (`pcx_to_clx.cpp:91-93`) and decodes
   per scanline into a `bytesPerLine`-sized buffer (`:126`, `:141-163`). PCX pads scanlines to an
   even byte count, so odd-width files mis-decode or error today.
3. **`transparentColor` unsupported.** C++ `PcxToClx` takes an optional transparent colour and
   emits transparent runs for it (`pcx_to_clx.cpp:94`, `:172-190`); cursors rely on this. Go has
   no such parameter.

## Acceptance criteria

- [x] Decoded rows are bottom-up exactly like `pcx_to_clx.cpp:169` (pinned by a test whose
      fixture has asymmetric top/bottom rows).
- [x] `BytesPerLine < width` returns an error naming the header field, and scanlines are decoded
      on `BytesPerLine` boundaries with padding skipped (test with an odd-width, padded fixture).
- [x] A transparent-colour parameter exists; pixels equal to it decode to the transparent index
      (test with a fixture containing that colour).
- [x] Existing synthetic PCX tests still pass, and a PCX fixture committed under
      `internal/testdata/gfx/` pins the corrected row order against C++ output.
- [x] RED first: capture the flipped-output and padding failures before fixing.

Refs: PR #3, devilGo-n89r

## Libraries
- none (stdlib only; PCX fixtures built in-test)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only brainer audit (pi session `devilGo-5wsb-beansaudit`, report `/home/blaze/.hermes/cache/scratch/pr3-beans-audit.md`); every cited anchor re-verified in the parent against the pinned devilutionX checkout (`ca91d778f`) before filing.


## Proof of Work (2026-10-28, PR #3)

- RED captured (worker): bottom-up `pcx_test.go:22`, frame loop `:36`, BytesPerLine error text
  `:51`, padding `:69`, row order `:87` — all failing before the fix.
- Fix: bottom-up emission (`pcx_to_clx.cpp:123,169`), `BytesPerLine < width` error naming the
  header field (`:91-93`), per-scanline decode with padding dropped (`:141-163`), variadic
  `DecodePCX(data []byte, transparentColor ...byte)` (matches `DecodeCEL` convention).
- Parent re-verification: gofmt/build/vet clean; 26 pkgs ok / 0 FAIL; `scripts/lint.sh` `0 issues.`.
- Commit `a2623b0`. Residual: real byte-goldens → `devilGo-l5fp`; multi-frame PCX not ported
  (stays single-frame, documented).
