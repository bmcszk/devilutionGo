# feat(render): text integration fixture parity — upstream pixel goldens

Cites bean: [Bitmap text renderer (CLX fonts)](.beans/devilGo-v1hy--bitmap-text-renderer-clx-fonts.md)

<!-- FILL after green: commit map, gate tails, review round -->

## What

Lands the last open acceptance criterion of the bitmap text renderer:
"Integration fixture outputs hash-equal (CI)".

- `internal/testdata/render/producer/textrender.cpp` — offline fixture
  producer, hand-compiled against the pinned upstream tree (ca91d778f),
  reproducing `test/text_render_integration_test.cpp`'s exact scenarios
  (26 fixtures) through upstream's own verbatim draw-path code. Never run
  by go test or CI (house producer discipline, see producer README).
- `internal/testdata/render/textfixtures/` — committed palette-index dumps
  + manifest (scenario, dims, sha256), produced by the above and
  cross-checked index-for-index against the pinned upstream PNGs.
- `internal/render/text_integration_test.go` — RED-first test rendering
  every scenario through DrawString and comparing all pixels + cursor.
- `internal/render/text.go` — minimal extensions the goldens exposed
  (cursor overlay, highlight range) with parity evidence.

## Fidelity & TDD

- RED run: failing assertions recorded in the lane log before any
  implementation (TDD program gate).
- Parity evidence: pixel-identical dumps, producer-vs-PNG proof table,
  plain-clone oracle run (both tails in bean PoW).
- Deviations from upstream behavior: <!-- FILL or "none" -->

## Gates

<!-- FILL: gofmt/vet/test/lint/beans-lint + plain-clone tail -->

## Review round

<!-- FILL: brainer verdict + findings + fix commits -->
