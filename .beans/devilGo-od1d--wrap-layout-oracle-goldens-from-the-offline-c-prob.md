---
# devilGo-od1d
title: wrap-layout oracle goldens from the offline C++ probe (text F4)
status: draft
type: task
priority: normal
created_at: 2026-11-11T17:30:00Z
updated_at: 2026-11-13T00:33:34Z
---

---
# devilGo-od1d
title: wrap-layout oracle goldens from the offline C++ probe (text F4)
status: draft
type: task
priority: normal
created_at: 2026-11-11T17:30:00Z
updated_at: 2026-11-11T17:30:00Z
---

# Wrap-layout oracle goldens (C++ probe run once)

## Symptom

text_test.go's wrap expectations (feat/text, devilGo-v1hy) are hand-derived
from committed glyph widths + the ported algorithm — reviewer F4: the bean's
wrap AC ("wrap widths identical for 20 sample strings") is not proven against
upstream OUTPUT, and the worker self-corrected the table twice before GREEN
(text.s2.log), the known failure mode of hand-derived evidence.

## Fix

Run the offline C++ producer chain ONCE (pattern: internal/testdata/render/
producer/oracle.cpp + the ADR-019 tools carve-out): a small gtest/probe that
feeds the 20 golden-source strings to upstream text_render.cpp DrawString/
WordWrap at the pinned commit and dumps per-line (startX, lineWidth) pairs
as TSV; commit the TSV under internal/testdata/assets/fonts/; a Go test
strict-loads it and asserts internal/render wrapLines matches line-for-line.

## Acceptance criteria

- [ ] C++ probe committed (source, like oracle.cpp) + 20-string TSV golden
- [ ] Go strict-loader test asserts wrapLines output == TSV per line
- [ ] beans-lint clean; no binary artifacts (TSV text only)

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Program-surface law (ADR-020/021 + one-tool law): no new binaries. Fixture
   instruments are subcommands of tools/goldens; user-facing surface stays
   exactly cmd/devilutionGo + cmd/devilutionGo-migrate. No `func main` outside
   the allowlist.
2. Tooling problems: fix immediately, no hacks.
