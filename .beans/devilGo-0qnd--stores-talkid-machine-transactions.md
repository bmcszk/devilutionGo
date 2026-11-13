---
# devilGo-0qnd
title: 'Stores: TalkID machine + transactions'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - stores
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-onlm
---

---
# devilGo-0qnd
title: 'Stores: TalkID machine + transactions'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - stores
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-01T19:28:10Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-onlm
---

# Stores

## Goal
All 24 TalkID states; buy/sell/repair/recharge/identify/premium/boy/healer; gossip texts.

## Research
stores.h:44-70; stores.cpp flows (StartStore re-entry :504,911-929,1192,1293); tests store_transaction/vendor/visual_store.

## Scope
internal/gameplay/stores: pure transaction core (prices, stock, premium refresh timers, gold checks) + UI-facing state struct the UI epic consumes.

## Acceptance criteria
- [x] store_transaction_test port green
- [x] Premium item slots 6 (D) / 15 (HF) refresh per the real C++ mechanism (town-entry refresh + in-place replace; no timer exists upstream @ pin)

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/coding-rules/SKILL.md — separate intent (transaction) from detail (UI refresh)
/home/blaze/.agents/skills/go-tests/SKILL.md — scripted store sessions

## Dependencies & blockers
devilGo-onlm.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:TalkID-count-23→24 + missing-dep-devilGo-rctk (+remap)


## IMPL NOTES (triage 2026-10-16)
1 root: TalkID transaction state machine; = devilGo-0qnd; runtime half of rctk's data bean.
2 local: beans show ✓; store sell = max(_ivalue/4, 1) stores.cpp:565-567 ✓; test/store_transaction_test.cpp exists ✓; BUT goal says "All 24 TalkID states" — stores.h:44-70 enumerates **24** (None…Barmaid, counted; rctk correctly says 24).
3 online: n/a
4 solution: fix 23→24; add missing dep devilGo-rctk (rctk: "Transaction state machine stays in devilGo-0qnd"; store catalogs/price rules/TalkID enum are rctk's) and devilGo-baxb (vendor stock calls SpawnSmith/…); remap devilGo-onlm→devilGo-onlm; S2.
5 understanding: game-layer transaction core consuming rctk price/stock tables + baxb/onlm generation; UI half in devilGo-glp0.
6 repro: beans show devilGo-0qnd; sed -n '44,70p' devilutionX/Source/stores.h | grep -c ,; sed -n '563,567p' devilutionX/Source/stores.cpp
7 fix: beans update devilGo-0qnd --blocked-by devilGo-onlm, then --blocked-by devilGo-rctk, then --blocked-by devilGo-baxb; body-replace-old "All 23 TalkID states" --body-replace-new "All 24 TalkID states".
8 validation: beans show shows 24 + three edges.
9 related: devilGo-rctk, devilGo-baxb, devilGo-onlm, devilGo-glp0
VERDICT: NEEDS-WORK:TalkID-count-23→24 + missing-dep-devilGo-rctk (+remap)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Count corrected (triage): TalkID states = 24, not 23.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

## Progress log
- 2026-11-02: worker1 (proc_ddcce3ffed89) died on STALE-PLUGIN transport after building the stores core green: TalkID machine 24 states, menus/pages/re-entry, hold+trade flows, price rules (sell/repair/recharge/boy), gold checks, buy rules, gossip mapping, 22 tests, 3629 lines. Salvaged uncommitted WIP as 37cd7a3 (worker had ZERO commits — chunked writes without committing). Remaining: premium refresh timers + full-repo gates -> continuation devilGo-0qnd-coder2 (proc_c8c8847712f3).

## Progress log
- 2026-11-02: LANE COMPLETE (proc_c8c8847712f3 exit 3 = transport, work done): 11 stores commits on dg-stores-talkid-machine, tip b04a29b. Premium refresh: worker PROVED no premium timer upstream (no PremiumItemCountdown in stores.{h,cpp} @ pin; refresh = InitStores clear + SetupTownStores on town entry + ReplacePremium in place) - owner re-verified against source; bean AC reinterpreted to the real mechanism, 6/15 slot counts + seeded tape + rotation all tested (13 premium tests). Plus StoreUp/Down cursor movers, Prior/Next page jumps, push-repeat counters. 35 pkgs ok, lint 0, tree clean.

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
