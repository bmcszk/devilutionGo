---
# devilGo-d12h
title: Convert lighttable fixtures .hex to human-readable TSV
status: completed
type: task
priority: normal
created_at: 2026-10-21T11:05:01Z
updated_at: 2026-10-21T11:12:01Z
parent: devilGo-qo09
---

# Convert lighttable fixtures .hex to human-readable TSV

Current `internal/testdata/gamedata/lighttables/*.hex` are xxd dumps — machine-readable only. Reviewer wants human-readable fixtures.

## Scope (3 slices, commit each)
1. **Converter** (~10 min): one-shot script (style of `scripts/fixture_fixtures.py`) decoding each `.hex` (xxd -r semantics) and emitting `.tsv`: header comment lines `# table: <name>`, `# dims: <rows>x<cols>`, `# source: devilutionX/Source/engine/light_tables.cpp:<lines>`, then decimal byte values, tab-separated. Delete `.hex` in the same commit. Sizes: falloffs 4096 B (16x256), cone 512 B, cathedral/crypt/nest/hell 256 B each (16x16).
2. **Decoder rewrite** (~10 min): replace `decodeHexDump`/`hexColumnWidth` in `internal/light/tables_test.go` with a TSV reader (skip `#`, parse decimal ints). Assert per-table byte counts from the TSV header, not hardcoded literals.
3. **Parity proof** (~5 min): before switching the decoder, capture decoded byte arrays from .hex; after, from .tsv; assert byte-for-byte identical (PoW lists the comparison result).

## Acceptance criteria
- [ ] `.hex` files gone; `.tsv` files committed with headers
- [ ] tables_test.go reads TSV; `go test ./internal/light/...` green
- [ ] Byte-for-byte parity proven in PoW

## Skills for coder
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/golang-testing/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only planner (pi-as-brainer session devilGo-plan9, /tmp/dg-plan9-report.md): 6 xxd .hex fixtures verified machine-only; TSV with dims/source headers = human-readable; decoder rewrite scoped to tables_test.go; byte-parity gate required. Note: planner byte-size estimates were stale — actual sizes measured at conversion time (falloffs 2048, cone 16384, others 4096).

## Libraries
- stdlib only (converter script); Go test decoder stdlib

## Proof of Work
- Superseded stale size notes (planner estimates): measured actuals used — falloffs 2048 B (16x128), cone 16384 B (16x1024), cathedral/crypt/nest/hell 4096 B (16x256).
- scripts/convert_lighttables.py (xxd -r semantics) -> 6 .tsv committed with # table/dims/source headers; .hex deleted.
- internal/light/tables_test.go: decodeTSV + parseDimsHeader replace decodeHexDump; TestDecodeTSV pins decoder.
- Parity: 6/6 byte-identical vs pre-change baselines (/tmp/dg-hex-baseline/).
- Parent gates: gofmt 0, build OK, vet OK, go test 547/28, golangci-lint 0 issues, just data-audit ok.
Residual: none.
