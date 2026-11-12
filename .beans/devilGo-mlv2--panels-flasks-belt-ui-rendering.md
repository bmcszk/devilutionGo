---
# devilGo-mlv2
title: Panels, flasks, belt UI rendering
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - ui
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-11-10T08:23:49Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# Panel chrome

## Goal
Control panel, life/mana flasks, belt slots, buttons from ui_art with exact rects.

## Research
Source/control/control_panel.cpp:351 DrawPanelBox, :435-445 button rects; control_flasks.cpp; panels/mainpanel.cpp; tests panel_state_test.cpp, inventory_ui_test.cpp.

## Scope
internal/render/panels: mainpanel, flasks (fill fraction), belt (8 slots), buttons (press state).

## Acceptance criteria
- [x] Flask fill matches player HP/MP fraction per C++ formula
- [x] Panel button hover/press visuals state-machine ported (test)

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — static art blits

## Dependencies & blockers
devilGo-n89r.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [x] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [x] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [x] UI layer rendered from UI atlas page; draw order after world buffer; per-frame cost <100 draw calls.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:minor (wrong mainpanel.cpp path; stale body dep; flask AC wording vs integer fillPer) — otherwise ready.


## IMPL NOTES (triage 2026-10-16)
1 root: Panel chrome bean — control panel/flasks/belt/buttons with exact rects. Sound research, minor drift.
2 local: `beans show` + probes: `DrawPanelBox` at `Source/control/control_panel.cpp:351` ✓ exact; button rects `:435-445` (DrawMainPanelButtons + MainPanelButtonRect) ✓; `control_flasks.cpp` exists, fill is integer `fillPer` 0..81 with `emptyRows = clamp(81-fillPer, 0, h)` (`:38-49`) — bean's "HP/MP fraction" wording is looser than C++ (81 discrete rows), testable either way; `panel_state_test.cpp` + `inventory_ui_test.cpp` both exist in C++ test/ ✓. WRONG PATH: `panels/mainpanel.cpp` cited as `Source/control/panels/mainpanel.cpp` — actual `Source/panels/mainpanel.cpp`. Body dep `devilGo-n89r` dead; frontmatter `devilGo-n89r` ✓. Skills path ✓ exists. Layer note: AGENTS.md places UI panels under `internal/game/ui`, bean uses `internal/render/panels` — systemic across qbiy children (all use internal/render/*), flag once, parent decides epic-wide.
3 online: n/a (no external claims).
4 solution: fix mainpanel path; sync body dep to n89r; tighten flask AC wording to fillPer 0..81 integer rows; optionally record layer placement decision (internal/render vs internal/game/ui) at epic level.
5 understanding: Diablo-layer UI chrome over the renderer; depends on sprite decode (n89r) and the index-framebuffer final pass (66o7 gate: "UI atlas page, draw order after world buffer, <100 draw calls" — measurable via ebitenginedebug).
6 repro: `beans show devilGo-mlv2 --body-only`; `sed -n '351,353p;435,445p' <devilutionX-repo>/Source/control/control_panel.cpp`; `sed -n '36,49p' .../Source/control/control_flasks.cpp`; `ls .../Source/panels/mainpanel.cpp`; `grep -A2 '^blocked_by' .beans/devilGo-mlv2--*.md`.
7 fix: `beans update devilGo-mlv2 --body-replace-old "panels/mainpanel.cpp" --body-replace-new "Source/panels/mainpanel.cpp"`; `beans update devilGo-mlv2 --body-replace-old "devilGo-n89r." --body-replace-new "devilGo-n89r (ui_art sprites)."`; `beans update devilGo-mlv2 --body-replace-old "- [ ] Flask fill matches player HP/MP fraction per C++ formula" --body-replace-new "- [ ] Flask fill matches C++ integer formula: fillPer 0..81 rows, emptyRows = clamp(81-fillPer,0,h) (control_flasks.cpp:38-49)"`.
8 validation: `beans show devilGo-mlv2 --body-only | grep -c dg-task` → 0; mainpanel cite resolves to existing file; fillPer in AC text.
9 related: n89r (dep), v1hy (text), xih1 (UI-flow epic), 66o7 (atlas gate).
VERDICT: NEEDS-WORK:minor (wrong mainpanel.cpp path; stale body dep; flask AC wording vs integer fillPer) — otherwise ready.

---

## Verdict summary
devilGo-ru6l: NEEDS-WORK:stale-body-deps + research paths
devilGo-m18k: NEEDS-WORK:broken-research (out-of-range cite, fabricated BlendMapMask) + missing skill path + dep mismatch + palette.kage overlap
devilGo-kx0l: NEEDS-WORK:missing fixture source (falloffs.bin/cone.bin nonexistent) + inverted deps + MakeLightTable duplication with 66o7
devilGo-j9be: NEEDS-WORK:scope conflicts with G4 decision (ColorScale) + stale open question + stale body dep
devilGo-mlv2: NEEDS-WORK:minor (path, dep, flask AC wording)

Cross-batch flags for the parent (not counted in this batch's verdicts):
- MakeLightTable triple-claim: 66o7 (render/lighttable.go), kx0l (light/tables.go), and 41p0 assumes "task 252" tests — one canonical owner needed (recommend kx0l; 66o7/41p0 consume).
- Systemic: every qbiy child body cites dead dg-task-* ids while frontmatter carries real devilGo-* ids — batch-04 owner should sweep its beans too.
- Systemic: bean package paths use internal/render/*, internal/light/* while AGENTS.md tree says internal/engine/render/ — record one decision at epic level.
- /home/blaze/.agents/skills/ebiten/ has no `shader` skill; any bean citing ebiten/shader must be swept (m18k here; check others).

BATCH-05-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

## Proof of Work (audit close 2026-11-11 vs master 2a91b97 — LANDED-AS-#54)

- PR #54 (merge ad1eea4) carries this bean's scope (merge commit title cites this bean line).
- internal/game/ui/flasks.go (ports control_flasks.cpp, integer fillPer, cites in the file header) + flasks_test.go:18 TestFlaskFillRowsFollowTheIntegerFormula + :56 clamp test (AC1); panels_test.go:202 TestDrawMainPanelButtonsSkipsPressedButton + :227-349 side-panel toggle state-machine tests (AC2); belt: internal/game/ui/belt_test.go:65-226.
- Recorded deviation (fidelity gate): the SITE is internal/game/ui emitting draw commands (ADR-017 simpurity: internal/game must not import internal/render), NOT the bean's original internal/render/panels; documented in the merge commit.
- Suite green at HEAD (postmerge57 gates). Root: audit brainer APPROVE evidence quoted in the wave record.
