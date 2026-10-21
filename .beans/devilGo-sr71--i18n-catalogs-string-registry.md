---
# devilGo-sr71
title: i18n catalogs + string registry
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - i18n
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-21T11:07:45Z
parent: devilGo-yjuv
---

# Translations

## Goal
.po→.mo pipeline + runtime lookup.

## Research
Translations/*.po (25); language.cpp:39-42; gotext-style loading.

## Scope
internal/engine/config/i18n with go:generate msgfmt step (or vendored .mo).

## Acceptance criteria
- [ ] en+pl 100-key parity test
- [ ] Missing key → English fallback (never empty)


## Libraries
- github.com/leonelquinteros/gotext — PO/MO catalog loading for i18n; pin v1.x

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-dev/SKILL.md — codegen for catalogs

## Dependencies & blockers
None.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DEFERRED — question answered by recorded milestone decision; revisit point documented in the question text above.
- bidi (sheen_bidi) deferred post-M2 (risk register).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: .po→.mo pipeline + runtime lookup with English fallback. Parent devilGo-yjuv, no
   blockers, no blocked_by ✓ consistent.
2 local: `ls Translations/*.po | wc -l` → **25 exactly** ✓; language.cpp:39-42 verified
   (.mo/.gmo extension handling, gotext-compatible loading premise) ✓; skills path exists ✓.
   Note: there is no `en.po` — English is the msgid source; "en+pl 100-key parity" is
   testable as source-string vs pl lookup.
3 online: n/a.
4 solution: none — ready. (Optional: clarify en = source-language comparison in AC.)
5 understanding: engine/config layer (gotext per AGENTS.md engine config responsibility);
   bidi (sheen_bidi) correctly deferred — sheen_bidi.hpp exists in C++ utils, port later.
6 repro: `ls <devilutionX-repo>/Translations/*.po | wc -l`; `sed -n
   '39,42p' Source/utils/language.cpp`.
7 fix: none.
8 validation: n/a.
9 related: devilGo-0tsf (settings 1:1 — distinct scope, no duplication), devilGo-yjuv.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

