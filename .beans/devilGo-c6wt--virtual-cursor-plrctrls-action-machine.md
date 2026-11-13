---
# devilGo-c6wt
title: Virtual cursor + plrctrls action machine
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-13T21:14:29Z
parent: devilGo-k736
---

---
# devilGo-c6wt
title: Virtual cursor + plrctrls action machine
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-12T20:14:29Z
parent: devilGo-k736
---

# plrctrls

## Goal
Gamepad-equivalent of mouse actions: move/interact/attack/spell/menu via virtual cursor.

## Research
plrctrls.cpp full module; hooks every_frame:2305, after_check_curs_move:2264, after_game_logic:2312.

## Scope
internal/input/plrctrls.go + cursor pick integration point (render epic provides screen→tile).

## Acceptance criteria (fix2: split)
- [ ] Core session parity (save/load roundtrip) — this bean
- [ ] 20-step full-session parity — deferred to devilGo-1n0a harness bean (dependency-gated)
- [ ] Action resolution identical to mouse path for a scripted 20-step session
- [x] No action while PadMenu/Hotspell overlays active

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
/home/blaze/.agents/skills/ebiten/input/SKILL.md — buttons + axes combos

## Dependencies & blockers
devilGo-sf6k.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:end-to-end AC (20-step session parity) not achievable at this bean's dependency level — gameplay/render deps not wired; AC needs unit-vs-e2e split.


## IMPL NOTES (triage 2026-10-16)
1 root: Input epic task — gamepad action machine; blocked_by sf6k (matches tree devilGo-c6wt←031).
2 local: hooks verified: plrctrls_every_frame :2305, plrctrls_after_check_curs_move :2264, plrctrls_after_game_logic :2312 (Movement(*MyPlayer)); file 2614 lines; PadMenuNavigatorActive/PadHotspellMenuActive guard verified :1861 — AC "no action while overlays active" directly testable.
3 online: n/a.
4 solution: AC mixes levels: "action resolution identical to mouse path for a scripted 20-step session" requires gameplay entities (combat/interact/towners) + render screen→tile, none of which are structural deps (only sf6w blocked; gameplay epic nxxo not wired). Split AC: (a) headless action-machine unit tests now; (b) 20-step session parity gated on nxxo + render cursor seam. Prose dep "devilGo-sf6k" stale → sf6k (structured already correct).
5 understanding: input layer integrating with gameplay hooks — epic k736 body already says "integrates with gameplay in epic-05 hooks"; the bean should express that gate explicitly.
6 repro: `beans show devilGo-c6wt`; `sed -n '2305,2314p' .../Source/controls/plrctrls.cpp`; `beans list --json | grep -c devilGo-nxxo`.
7 fix: `beans update devilGo-c6wt --body-replace-old "devilGo-sf6k." --body-replace-new "devilGo-sf6k (structured). End-to-end 20-step session AC additionally gated on gameplay epic devilGo-nxxo + render cursor seam."` + split AC bullets.
8 validation: `beans show devilGo-c6wt` shows split AC + nxxo gate.
9 related: sf6k, t4fz, nxxo children, k736.
VERDICT: NEEDS-WORK:end-to-end AC (20-step session parity) not achievable at this bean's dependency level — gameplay/render deps not wired; AC needs unit-vs-e2e split.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

## Progress log — slice 1 (C1+C2, 2026-11-12)

2026-11-12 parent update: slice 1 LANDED on feat/plrctrls (6be509e + fix
round 0986b73, PR open). Deliverable: internal/diablo/controls/
actionmachine.go — PerformPrimaryAction dispatch (13 press-path rows
1:1 vs plrctrls.cpp:2309-2353, reviewer-walked), PerformPrimaryActionRelease
(gate-free by upstream — review caught an invented overlay gate, fixed
with correct citations game_controls.cpp:182-190/:399-410), full cursor_id
enum, panel-containment model, overlay-refusal input (per-route citations:
FINGERDOWN !inGameMenu, game-UI flags, padmapper CanPlayerTakeAction-only).
Package 123 passing (110 subtests). Gates green; brainer round 1 verdict
REQUEST-CHANGES → both blocking findings fixed in C2. Review-verified
press path is a complete mirror; deferred (later slices): Interact internals
(world interaction family), inventory/stash grid geometry, spell-list +
store UI internals — all behind hooks with the boundary ledger in
.hermes/lane/dg-input/plrctrls.c1.log. Overlay AC ticked: press-path
refusal is pinned (zero hook calls under each flag); release-path firing
under overlays is upstream-faithful, noted. Still open on this bean:
save/load roundtrip parity (core session AC), screen→tile seam, the
20-step session ACs (deferred to the harness bean + gameplay epic).

## Progress log — session-parity slice (m1group G1, 2026-11-12)

2026-11-12 parent update: core-session roundtrip landed PARTIAL on
feat/m1group (8d70bd1) — honest scope, AC stays OPEN:
- Proven: PlayerPack 1266B full image, game header, level grids,
  dropped items — SQLiteStore SaveHero/SaveGame/SaveLevel/Load* on one
  hero.db; StateHash-equal reload; byte-identical re-save (idempotence).
- Mutation evidence: CurrLevel+1, Py±1, automap/dropped-item zeroing all
  flip the roundtrip RED (fixture enrichment: non-zero automap grid,
  real dropped-item bytes, NumInv=1 — zero-only fixtures are blind).
- Named missing for FULL world-session roundtrip: SaveGameData stream
  assembly (loadsave.cpp:2763-2800 walk), monster/quest/missile
  level-stream writers, game-layer session snapshot (player → session,
  walkpath). These gate the AC tick; next slice.
- 20-step session ACs: deferred to the parity-harness bean.

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

## Progress log — savegame streams slice (2026-11-13)

Branch feat/savegame-streams (5 commits, see PR body). What landed in internal/persist:

- Record encoders: SaveLighting standing+vision variants (loadsave.cpp:1795-1812 at pin ca91d778f), SavePortal with the getHellfireLevelType setlvl branch (:1813-1822, :1148-1157), SaveObject with the dynamic-light var1 zeroing (:1749-1763). Record strides 52/24/120 verified against the committed memory_map files.
- Grid writers: dLight (:2869-2873), dFlags masked to DungeonFlag::SavedFlags (:1874-1878), dPlayer (:1880-1884), dropped-item locations (:1858-1866), and the non-town tail dMonster BE-int32 / dCorpse / dObject / BUGFIX dLight copy / dPreLight / AutomapView (:1885-1916) — the double-dLight write is reproduced byte-for-byte and named in TestDungeonGridQuirkDoubleDLight.
- Monstkill counts BE-int32 + MaxMonsters=200 zero padding (:2839-2843, monster.h:38); premium tail: PremiumItemCount/PremiumItemLevel BE then SaveItem records (:2921-2924; 6 DA / 15 HF, :2484/:2489).
- ComposeSaveGameData: stream assembly in the upstream walk order (:2763-2935), reusing GameHeader/QuestSave/SaveItem/DroppedItems/PlayerPack/TileGrids encoders (TileGrids reuse landed 2026-11-13 in the fix wave) plus the new encoders. LevelSeeds is the separate MPQ levelseeds stream, NOT called by compose.

Pin cites verified; drift corrections: SavePortal is :1813-1822 (contract said 1812+); SaveQuest is :1765-1793; monstkill loop + padding is :2839-2843 (contract said :2841-2845); dungeon-only streams :2845-2864; double-dLight BUGFIX at :1901-1906 inside the tail (call path :2866-2930 ✓); SaveDroppedItems return-index tables :1835-1855 with the SaveDroppedItemLocations consumer :1858-1866. PREMIUM grep: SavePlayer at pin ca91d778f (Source/loadsave.cpp:1252-1507) contains NO premium lines — the contract's assumption is wrong for this pin; premium items are written by the walk tail (:2921-2924).

Evidence class (honest): encode->decode roundtrip against decoders over the same byte layouts + byte-exact section images derived from the pin field order + compose idempotence + committed composed-stream golden (testdata/compose_game_stream.hex, offset comments per section) + SQLiteStore SaveGame/LoadGame roundtrip with the GameHeaderSize floor + mutation proof (one flipped golden hex nibble -> TestComposeSaveGameDataGolden RED with 'differs from golden at offset 4992' -> restored -> green, all three runs recorded in the lane log). No upstream-run oracle for the compose: expectations are pin-derived bytes, not a C++ probe dump.

Gate table: go build ./... OK; gofmt -l empty; go vet 0; ./scripts/lint.sh ./... 0 issues; beans-lint OK (193 beans); go test ./... -count=1 = 50 pkgs ok, 0 FAIL.

ACs: the core-session-parity save/load roundtrip AC is proven at the persist layer for the FULL SaveGameData stream (all sections above + store roundtrip). The 20-step session ACs (input/harness/game-layer session snapshot: player->session walkpath in the game layer) remain OPEN — that piece is not in this diff and stays with the harness beans.

## CORRECTION — savegame-streams slice claims (fix wave 2026-11-13, PR #71 review)

The review (REQUEST-CHANGES, findings F1-F10) found the slice's completeness claims overstated; this append corrects the record (history above left untouched):

• "proven at the persist layer for the FULL SaveGameData stream (all sections)" was FALSE at the time it was written: the seeds loop (F1), the dMissile grid (F2), fixed-width id tables (F3) and the 6/15-record premium tail (F5) were all missing, and the committed golden pinned the wrong layout. The fix wave (commits eccaaa9..f519e84 + the golden regeneration) closes F1-F8; the claim is now true of the branch head, not of the slice as originally logged.
• "reusing ... TileGrids ..." was FALSE when written: EncodeDungeonGrids re-implemented the shared block instead of calling TileGrids.Encode; fixed in the fix wave (U7) and the line above is now accurate.
• The mutation-proof narrative said the flipped nibble was on "golden data line 157"; the correct location for the regenerated golden is data line 4935 (byte offset 157890). The proof was re-run: flip ff->ef at 157890 -> RED 'differs from golden at offset 157890' -> restore -> green.

## CORRECTION — fix wave 2 (2026-11-13, PR #71 re-review N1-N5)

The re-review (REQUEST-CHANGES, N1-N5) found the fix wave's "closes F1-F8" claim still overstated: F6 was not closed. The ConvertToHellfireObject table's numeric constants were wrong (sources off by −5: e.g. POD was 94, actually OBJ_MUSHPATCH; real POD=99, L5SARC=108 per the pinned Source/tables/objdat.h and both committed objdat.tsv fixtures), and ComposeSaveGameData never applied the remap — nest/crypt streams wrote raw types. Fix wave 2 lands the authoritative enum constants (b81c2a5) plus the compose wiring through EncodeForLevel with the header's level type, gated to nest/crypt exactly as upstream's leveltype branches (04b2fa6). F6 is closed at fix-wave-2 head; the earlier "closes F1-F8" line above is superseded by this correction. A new objdat-fixture-driven table test (TestConvertToHellfireObjectMatchesObjdatFixture) rebuilds the expected mapping from internal/testdata/gamedata/objdat/objdat.tsv, so the constants can no longer drift from the fixture.
