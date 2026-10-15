---
# devilGo-i56j
title: devilutionX → Go+Ebiten+SQLite whole-game conversion
status: draft
type: milestone
priority: normal
created_at: 2026-10-15T15:59:25Z
updated_at: 2026-10-15T21:41:42Z
---

# devilutionX → Go + Ebitengine + SQLite — WHOLE-GAME conversion

## Goal (binding, user 2026-10-15)
Rewrite devilutionX (C++ Diablo 1+Hellfire engine, ~/github/bmcszk/devilutionX, ~134.5k LOC Source/) as a complete, playable Go game in ~/github/bmcszk/devilutionGo using **Golang + Ebitengine + SQLite**. This epic is the session goal; all work hangs under it.

## Requirements (verbatim user directives)
- Whole game conversion — not a demo: gameplay, levels, entities, UI, saves, networking, audio, Hellfire.
- Reverse engineering first: gather ALL requirements with details, examples and know-how from the C++ tree before coding.
- Plan is the most important step: identify EPICS → tasks → subtasks, every bean with all required sections (research w/ citations, scope, acceptance criteria, skills for coder, dependencies, open questions).
- Method: keep-on-going discipline (this epic = the one active epic; draft→triage→todo→claim→PoW→close), pi brainer/RPC for research and triage, delegation per dev/pi-delegation.

## Global acceptance criteria (epic closes only when ALL hold)
- [ ] Requirements study completed and committed to devilutionGo/docs/ — all 12 subsystem areas, every claim cited file:line or URL, ≥100 distinct devilutionX source paths cited.
- [ ] Full beans tree created under this epic; every coding leaf carries `## Skills for coder` with existing paths; blockers wired.
- [ ] CI in devilutionGo: build + vet + golangci-lint + tests green on every push.
- [ ] Determinism harness: same-seed dungeon parity tests vs C++ fixtures pass.
- [ ] Spawn (shareware) vertical slice playable: title→new game→town→dungeon L1→combat→death/level-up→save/load via SQLite.
- [ ] Full single-player game completable through Diablo (all 4 dungeon tiers + quests).
- [ ] Hellfire content (crypt/hive, monk, new items/spells) playable behind the same binary.
- [ ] Multiplayer: 2 clients over TCP complete a shared session (CMD_ protocol port).
- [ ] Save/hero files: SQLite persistence, roundtrip load identical game state (hash-verified).
- [ ] Timedemo-style deterministic replay regression green in CI.
- [ ] Audio: sfx + music play through Ebiten audio pipeline.

## Process (binding)
1. PLAN FIRST — research contract dispatched via pi RPC (done at epic creation); deliverables: docs requirements study + full bean tree; orchestrator validates structurally before any promotion.
2. TRIAGE WAVES — each draft triaged by research subagent (pi/brainer) per beans skill contract before todo.
3. CLAIM LOOP — `beans list --ready` heartbeat; four-pass leaf work; PoW in every bean before completed; epic closes only with all children terminal + comprehensive e2e PoW.

## Research
- Local: devilutionX tree (Source/ 432 files), docs/BACKGROUND.md, test/ fixtures; sizes measured 2026-10-15 (see contract /tmp/dg-plan-contract.md).
- Online: ebitengine.org performance tips; pkg.go.dev OpenDiablo2/mpq + suprsokr/go-mpq (PKWARE DCL read-only, Huffman gap risk); mattn/go-sqlite3 vs modernc.org/sqlite; ebiten audio 48kHz stereo float32.
- Full study lands in devilutionGo/docs/requirements.md (from this epic's phase-0 task).

## Skills for coder
- (tree-level; leaf beans each carry their own) go-dev, coding-rules, ebiten/* sub-skills, beans.

## Dependencies & blockers
- None (root). Everything else blocks on or hangs from this epic.

## Open questions
- Save compat: import C++ saves or fresh-format-only? (decide in phase-0 with evidence)
- MPQ path: reuse retail archives via Go reader vs pre-extracted assets — Huffman gap decides.
- Lua scripting: gopher-lua embed vs drop-for-v1.



## Process log
- 2026-10-15: Plan wave 1 complete via pi RPC (_glm_first): /tmp/dg-requirements.md (115K, 12 subsystems) + 65-bean tree — validated 448 structural gates, imported as milestone(1) + 13 epics + 46 tasks. Bean mapping: /tmp/dg-id-map.json.
- 2026-10-15: USER DIRECTIVE — totality required: every menu item, NPC, cow, pickable item. Deep-extraction wave 2 dispatched (Volume II inventories V1-V12 + Phase-7 data-domain tasks dg-task-1NN).
- beans CLI hierarchy (this build): milestone→epic→task ONLY; nothing under tasks; sub-10-min work = AC checklist items inside tasks. Importer lesson recorded.
- Deliverables land in-repo via phase-0 pi contracts (docs/requirements.md = Volume I+II), not by orchestrator hand.



## FIDELITY MANDATE (user, binding 2026-10-15): 1:1 CONVERSION
- Every ported system is 1:1 with devilutionX C++: identical algorithms, constants, formulas, RNG behavior, data values; byte-level save/MP compatibility; pixel-identical rendering output — subject only to deviations explicitly recorded as decisions in the owning bean + requirements doc.
- NO approximations, no 'inspired-by', no modernized rewrites. A deviation without a written decision entry = defect.
- Parity evidence is MANDATORY in every implementation task's PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table — chosen per domain.

## TDD MANDATE (user, binding)
- Test-first: the FIRST acceptance criterion of every implementation task is a failing test (golden / parity / unit) written before implementation; PoW includes the test run output.
- Domain forms: data = row-count + spot-field tests; graphics = golden-image/pixel-diff; logic = C++ fixture parity tests; net = packet hexdump roundtrip; saves = .sv byte-compat roundtrip.

## WORKER SKILLS MANDATE (user, binding)
- Every coding task carries ## Skills for coder with real, on-disk paths (validator-enforced at import; retrofit pass fills gaps). Worker dispatches pass skill PATHS + binding rules as checkable gate items.



## Process log (cont.)
- 2026-10-15 ~19:35: NETWORK OUTAGE killed extract+graphics+sweep runs (HTTPS egress dead, ICMP OK; omniroute 502/503 cluster). Lesson: RPC client surfaced empty/garbage text as success — driver patched: persist batch only when entries>=85% and marker present; resume gate by entry count.
- Wave 2b GRAPHICS 1:1 DONE+VERIFIED: /tmp/dg-graphics.md (G1-G5, 1218L, 348 cites) + /tmp/dg-graphics-part2.md (G6-G11, 1075L, 261 cites) incl 6 EXTRA TASKS → imported devilGo-gk5e/66o7/7aul/b782/qeyu/l5fp with blockers wired.
- Wave 2 VOLUME II RECOVERED: pi had left completed generator at /tmp/gen_vol2.py; orchestrator repaired syntax/runtime bugs (/tmp throwaway, counted exception) → ran: /tmp/dg-requirements-append.md 319,720B / 3,824L, all V1-V12 verified (168/83/95/90 items rows in tables, 100 uniques, 114 missile rows...). Phase-7 bean wave re-dispatched (proc_ae3627dfd5c9).
- Wave 3 sweep RELAUNCHED at 2 lanes (proc_908a3be103e9), corrupt batches deleted, valid re-detected by entry count.


- 2026-10-15 20:30: Phase-7 imported: 20 data-completeness tasks dg-task-101..120 (pi generator recovered from /tmp/w/gen.py, %s-format bug fixed, validated all JSON/parents/blockers/skills, 45 blockers wired). Ledger now 72 tasks. TDD+1:1 mandate present in ALL task bodies (46+26 retrofit verified). Volume II errata captured: class rows 1080 not 1122; AiProc 40 not 41 (dg-task-119 tracks correction). Sweep relaunched clean single-driver (2 lanes, marker-OR-content gate, vol3-001/002 preserved).


## Graphics engine → Ebiten migration knowledge (COMPLETE, verified)
- `/tmp/dg-graphics.md` (G1-G5: frame architecture, level tiles, sprite formats CEL/CL2/CLX/PCX byte-level, palettes+TRN, lighting) + `/tmp/dg-graphics-part2.md` (G6-G11: entity/UI sprite plumbing, text/fonts, effects, asset manifest, consolidated Ebiten mapping decisions G10-D1..Dn, pixel-parity test plan G11).
- 609 distinct source cites; online-verified Ebiten facts (no examples/palette; batching rules; 4096×4096 atlas; SetCursorShape; Kage pixel-vs-texel units).
- G10 decisions binding for the port: D1 640×480 fixed Layout + FilterNearest; index framebuffer + 16 palette-row textures + Kage palette-lookup shader (NOT ColorScale); atlas pages ≤4096²; batching = same atlas+blend+filter runs (count with -tags=ebitenginedebug); Update/Draw split with ProgressToNextGameTick interpolation; CPU-side palette fade/cycle only via ReplacePixels on palette textures.
- These decisions are enforced per-task via 'Ebiten graphics-migration gates' acceptance criteria retrofitted into all 12 graphics beans (11 graphics + decoder task).
- Sweep (vol3 per-file, proc_d67e137846a7) is filling the remaining detail (per-file requirements incl. every renderer file).


## Volume III sweep COMPLETE (2026-10-15 23:24)
- ALL-BATCHES-COMPLETE: 52+4 batches, 668/670 per-file entries (2 uncovered: Source/DiabloUI/dialogs.h + selstart.h — trivial header pair, covered by selok.h entry's context).
- Merged docs deliverable committed+pushed: commit f676095 on master (repo created: github.com/bmcszk/devilutionGo), 2.6 MB docs/requirements.md = Vol I + II + III + Graphics G1-G11.
- Bean ledger committed (.beans/ 86 files).
