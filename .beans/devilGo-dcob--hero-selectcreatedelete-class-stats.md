---
# devilGo-dcob
title: Hero select/create/delete + class stats
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-25T13:41:19Z
parent: devilGo-xih1
blocked_by:
    - devilGo-xoap
---

---
# devilGo-dcob
title: Hero select/create/delete + class stats
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-25T13:40:48Z
parent: devilGo-xih1
blocked_by:
    - devilGo-xoap
---

---
# devilGo-dcob
title: Hero select/create/delete + class stats
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-25T12:12:09Z
parent: devilGo-xih1
blocked_by:
    - devilGo-xoap
---

---
# devilGo-dcob
title: Hero select/create/delete + class stats
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-29T15:29:47Z
parent: devilGo-xih1
blocked_by:
    - devilGo-emdz
    - devilGo-pu3r
---

# Hero screens

## Goal
selhero over persistence (list/create/delete), class stat preview.

## Research
DiabloUI/hero/selhero.cpp; _uiheroinfo diabloui.h:60-72; callbacks :81,93; pfile_ui_* API pfile.h:120-125; MAX_CHARACTERS 99 pfile.h:24.

## Scope
internal/ui/selhero.go wired to internal/persist hero store interface (stub until epic-08).

## Acceptance criteria (fix2 additions)
- [x] D7 six-class packs from data: replace the literal bootPackSource (device_lane.go:103-106) with packs built from the committed per-class attributes.tsv (+ starting_loadout.tsv via txtdata.LoadClassLoadout) for all six classes; Bard/Barbarian refused at the gate with a RECORDED deviation note (upstream gates the list by assets/Hellfire — class list selhero.cpp:170-214 @e1ad0c2b6, gates :179/:183/:187)
- [x] D8 preview parity row-exact for what the screen shows: six-class x {level=1, baseStr, baseMag, baseDex, baseVit} table test against committed expected values transcribed from pfile_ui_set_class_stats (pfile.cpp:724-732 @e1ad0c2b6) + the six attributes.tsv files
- [x] D2 class ordinal: pickClass submits the picked CLASS ORDINAL (classdat row order), not the menu index; a test with a middle class row missing proves index != ordinal still submits the right class word
- [x] D3 no silent overwrite: creating while a hero exists never replaces a row — the new hero takes the first free slot (persist.HeroSlots.FirstFreeSlot exists: internal/persist/pfile.go:131-140); test: second create keeps the first hero
- [x] D4 99-cap: with all 99 slots occupied the create is refused and surfaced (no kernel submit); store-side bound already exists (validateHero, MaxCharacters=99) — pin the screen side
- [x] D5 delete confirm: Delete on a focused hero opens Yes/No; only Yes calls DeleteHero; No/Escape deletes nothing (upstream selhero.cpp:613-624 @e1ad0c2b6)
- [x] D6 delete disabled with no hero focused: pin the no-op with a test (matches upstream :139-153/:557 @e1ad0c2b6)
- [x] D10 (RF1) created heroes settle in the shipped binary: a windowless test driving windowGame.Update (NOT the test boot's frame) sees the store row after create + N Updates; fix the front-screen early-return (window_ebiten.go:171-181) or settle the create on close — highest-value item in the lane

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
/home/blaze/.agents/skills/ebiten/input/SKILL.md — text input for name (16 chars, _uiheroinfo.name)

## Dependencies & blockers
devilGo-emdz.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [x] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [x] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale devilGo-emdz prose ref.


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic (xih1) task — selhero screens over persistence; structurally parented + blocked_by emdz (front screens), mirroring tree devilGo-dcob←028.
2 local: `beans show` OK; `_uiheroinfo` struct verified DiabloUI/diabloui.h:60-72 (name[16] etc.); callbacks :81/:93 (gfnHeroInfo, UiSelHeroSingDialog) verified; `MAX_CHARACTERS 99` pfile.h:24 verified; `pfile_ui_*` cluster pfile.h:120-128 verified; hero/selhero.cpp exists (22,830 bytes); `playerdat` = tables/playerdat.cpp exists.
3 online: n/a (no external claims).
4 solution: AC incomplete vs totality directive: selhero.cpp:437-463 also defines Bard (asset/test-gated) and Barbarian (test-gated) class entries — AC "warrior/rogue/sorcerer(+monk HF)" silently drops them; record decision or add them. No domain-specific parity AC (e.g. name-length validation = UiValidPlayerName, class-stat defaults row-exact vs playerdat). Prose dep "devilGo-emdz" stale → emdz.
5 understanding: UI layer (internal/ui/selhero.go) consuming persistence store interface — correct seam; stub-until-epic-6xu2 noted in scope.
6 repro: `beans show devilGo-dcob`; `sed -n '60,72p' ~/github/bmcszk/devilutionX/Source/DiabloUI/diabloui.h`; `sed -n '437,463p' .../DiabloUI/hero/selhero.cpp`; `grep -n MAX_CHARACTERS .../pfile.h`.
7 fix: `beans update devilGo-dcob --body-replace-old "devilGo-emdz." --body-replace-new "devilGo-emdz (structured blocked_by already set)."` + append AC line for Bard/Barbarian handling + UiValidPlayerName parity.
8 validation: `beans show devilGo-dcob` shows updated dep + new AC item; re-grep selhero.cpp:183-187 gates.
9 related: emdz, vf8q, h6wn (menu spec), xwtt (pfile API), 0tsf.
VERDICT: NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale devilGo-emdz prose ref.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

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

## Triage hl1 (2026-11-25, brainer task-hl1-triage, parent-verified against 949c728)

Verdict: NEEDS-WORK — ACs re-scoped below (D1-D8, D10; D9 split to its own bean).
- RF1 (inherited from the #81 re-review, confirmed live): a created hero never settles in the shipped binary — window_ebiten.go:171-181 stops the lane while front screens are up and no screen exists past FrontSelHero (internal/present/scene/frontflow.go:11-16); windowless tests pass because the test boot drives frame() itself. D10 owns the fix.
- Slot-0 overwrite (unnamed defect, found this run): pickClass always writes Slot: 0 (selhero.go:341) — a second create silently replaces the first hero (slot UNIQUE upsert, internal/persist/queries/store.sql:5-10). D3 owns the fix.
- Upstream cite fix: class list is selhero.cpp:170-214 @e1ad0c2b6 (gates :179/:183/:187), not :437-463.
- Load path: upstream SelheroListSelect offers Load Game / New Game (selhero.cpp:217-236 @e1ad0c2b6); Go maps a hero row to the CREATE flow (selhero.go:205-215). Split to the D9 bean (blocked-by this bean).
- D1 DECISION (parent, recorded): ADR-003 "one hero = one .db file" (docs/adrs/adr-003.md:12) STANDS — per-hero single_N.db files are the model; the boot's single multi-row single_0.db (window_ebiten.go:49-54) is a recorded demo-wave simplification whose migration rides the D9 bean. D3/D4 slot policy implements against the persist store seam so the migration does not rewrite it.
- Edges: blocked-by emdz REMOVED (satisfied by the landed uuyh front-screens slice); blocked-by pu3r REMOVED (pu3r completed 2026-11-13 — dependency satisfied); blocked-by xoap ADDED (the F6 delete pin must be re-pinned before D5/D6 touch deleteFocused). Full triage: .hermes/lane/m2-hero-load/result.md (gitignored lane evidence).


## Summary of Changes

D1 recorded (ADR-003 per-hero single_N.db model stands; boot single-db migration deferred to the load bean). D2: class ORDINAL submitted, not menu index. D3: first-free slot — a second create no longer overwrites the first hero. D4: 99-slot refusal before any kernel submit. D5+D6: delete asks Yes/No (Enter only), idle KeyUnknown frames ignored, delete-without-focus pinned no-op. D7: starting packs built from the committed per-class attributes.tsv for all six classes; vitals in the 1/64 raw scale upstream stores (warrior 4480); Bard/Barbarian refusal = recorded asset-gate deviation. D8: preview parity against a hand-transcribed six-class table (per-row attributes.tsv cites). D10 (RF1): a created hero settles in the shipped binary — the window loop resumes the lane behind the front screens (clock now runs from the first frame; behavior change recorded above).

## Proof of Work

- RED evidence per leaf recorded in the round logs (highlights): D2 `CreatedPack class = 1, want 2`; D3 `store rows after two creates = 1 ([{0 Second Hero}])`, want 2; D4 `CreatedPack events = [...], want none` at 99 slots; D5 idle-frame RED `idle KeyUnknown frame answered the confirm: deleteConfirmFor cleared, want it still armed`; D7 round-1 repair RED `HPBase:70 ... want ... 4480`; D10 `store rows after create + 10 Updates = [], want 1`. D6 honestly PIN-ONLY (pre-existing guard covers it).
- Commits: fa90e60 (D2), cb83d6f (D3+D4), 9e4a10e (D5+D6), 3d9693b (D7+D8 first pass), 53e6b27 (D10), plus review-round-1 repairs eccd5d1 (F1+F2+F4+F9), 002a41d (F3), 8ea4ac3 (F6+F7), 42cc541-parent (F10 docs).
- Independent review: round 1 REQUEST-CHANGES (F1 vitals scale, F2 self-comparing parity gates, F3 idle-frame confirm — all parent-verified against the tree, then repaired); round 2 APPROVE (reviewer re-derived all six parity rows from the committed TSVs independently and verified upstream Fixed26_6 semantics).
- Parent gates at branch head 42cc541: `go test -count=1 ./...` all ok; `go vet ./...` rc 0; gofmt clean; lint 0 issues; windowed cmd suite ok with real archives (AC4 included); windowed scene fails only on the pre-existing TestScreenshotCaptureWriter panic (own draft bean).
- Residual: D9 (hero-row Load Game / New Game over a production SessionArchive) split to its own draft bean (blocked-by this one); the StartingPack loadout extension is deferred with it (recorded above); D8's level column is a documented constant-1; the single-db boot enumeration migrates in the load bean.



The three Fidelity & TDD gate boxes above are ticked by this wave's evidence: RED-first recorded per leaf (round logs), parity evidence row-exact committed (transcribed six-class table, reviewed round 2), and the only behavior deviation (Bard/Barbarian asset gating) is recorded in this body.
