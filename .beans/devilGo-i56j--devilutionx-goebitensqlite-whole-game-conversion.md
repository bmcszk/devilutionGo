---
# devilGo-i56j
title: devilutionX → Go+Ebiten+SQLite whole-game conversion
status: draft
type: milestone
priority: normal
created_at: 2026-10-15T15:59:25Z
updated_at: 2026-10-16T14:27:07Z
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


## Commit: lighting + sound specs + 8 lighting beans (2026-10-16)
- Commit 7897477 pushed to master
- docs/lighting.md (L1-L8: state, tables, sources, render, vision, net, impl, risks)
- docs/sound.md (SN1-SN9: arch, math, playback, streaming, catalog, options, constraints, parity)
- 8 beans: devilGo-kx0l,bodd,pdpc,m18k,9o2h,h8uq,41p0 (+253 mapped to pdpc)
- All blockers wired (14 cross-refs to existing graphics/decoder tasks)


## DECIDED (grilling session 2026-10-16, user-confirmed)
1. SAVES: SQLite is the only live save format. One-time CLI importer tool (`devilutionGo-import old.sv` → sqlite.db) preserves legacy .sv/.sp compatibility; byte-compat importer tested as proof. No dual-format live writes.
2. MPQ: Build-time extraction to loose assets for v1 (per AGENTS.md). Compression gain marginal; runtime MPQ reader = deferred optional post-1.0 feature bean (NOT cut).
3. LUA: Deferred post-1.0. Pure-Go constraint project-wide → gopher-lua when implemented (no CGO). The LuaEvent hook points get stubbed in engine (0 VM) so future enablement is drop-in.
4. ORDER: Epic order = natural dependency order (Phase 0 scaffold first), no artificial ordering constraints.
5. RESOLUTION (1:1 correction): devilutionX renders dynamically at gnScreenWidth/Height (any window size, zoom toggle). Go port = fixed 640×480 logical world surface, integer-scaled to window. Parity harness tests 640×480 AND a 2x window. G10 fixed-640 note corrected.
6. DELEGATION (HARD): ALL coding beans execute via pi RPC sub-agents (contract per bean: read bean body → implement → tests green → PoW on bean). Orchestrator does docs/beans/review only — no direct code edits. Batch size = 1 epic.


## SAVE ARCHITECTURE AMENDMENT (user decision 2026-10-16)
- **Per-hero SQLite DB**: every save = separate SQLite database file (one .db per hero/game), mirroring the C++ per-hero .sv file model.
- **Global SQLite DB**: one additional shared database for cross-hero state (stash/shared storage, global settings, metagame).
- Importer writes into this model: legacy `single_0.sv` → `single_0.db` (per-hero), stash file → global.db.
- Persistence epic (devilGo-6xu2) tasks amended accordingly.


## SETTINGS FORMAT AMENDMENT (user decision 2026-10-16)
- Settings stored 1:1 to devilutionX: `diablo.ini` INI file (same path convention via SDL pref path `~/.local/share/diasurgical/devilution/`, portable.ini support), read/written by a Go INI codec matching devilutionX's utils/ini.cpp semantics (sections, keys, comments preserved).
- Global SQLite DB does NOT hold settings — it holds stash/metagame only.
- INI keys/values/sections byte-compatible so a user can copy diablo.ini between devilutionX and devilutionGo.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## FULL LEDGER TRIAGE COMPLETE (2026-10-16, pi-as-brainer wave)
- 93/93 beans triaged one by one, 10 batches + 1 gap-fill batch, evidence spot-checked by parent.
- Verdicts: 36 TRIAGED-OK, 56 NEEDS-WORK, 1 DECISION-NEEDED (devilGo-lfvi: package layout AGENTS.md-vs-plan — resolved by AGENTS.md three-layer rule).
- Write-back: all 93 beans carry TRIAGE 2026-10-16 notes with verdicts.
- 5 root-ready tasks promoted to todo; 21 TRIAGED-OK tasks stay draft (blocked by draft deps).
- NEEDS-WORK fix classes: 23 dep-wiring, 5 overlap seams, 7 decision-alignment (MPQ/ini/Lua/resolution amendments), 5 count corrections, 1 cite fix, 8 AC fixes + 5 lighting batch fixes (fixtures missing, clx_render cite range, MPQ-vs-decision conflicts).
- Batch outputs: /tmp/dg-triage-out/batch-01..10.md. Next: fix pass over NEEDS-WORK beans, then re-verdict.


## IMPL NOTES (triage 2026-10-16)
1 root: Root milestone owning the whole port; parent of all 13 epics (verified: 13 children via GraphQL). Carries binding fidelity/TDD/skills mandates and all six grilling decisions + two amendments (saves, settings).
2 local: `beans show` full body read; decisions present verbatim (SQLite-only+importer, extract-at-build, Lua deferred post-1.0 w/ gopher-lua, fixed 640×480 scaled, pi-RPC hard delegation, per-hero DB + global stash DB, diablo.ini 1:1). Commit trail in devilutionGo confirms amendments landed as docs (git log: 71b17dd, c879189, fde12b2). docs/requirements.md committed (2.6 MB, per log; ls docs/ ✓).
3 online: n/a (external claims live in child epics; ebitengine/go-mpq facts were verified during plan wave per process log).
4 solution: none structurally. Optional: archive the /tmp/dg-* references (many are ephemeral) into docs/ or note which are reproducible; process log already records commits.
5 understanding: The ledger's constitution — every epic/task inherits its mandates and must not contradict its DECIDED section (this batch found 3 epics that do: 5wsb, 6xu2, yjuv).
6 repro: `beans show devilGo-i56j --body-only | head -40`; `cd devilutionGo && git log --oneline -5`; `beans query --json '{ bean(id: "devilGo-i56j") { children { id } } }'`
7 fix: none required. Optional: `beans update devilGo-i56j --body-append "- Note: dg-epic-NN legacy labels in epic bodies map via docs; swap to devilGo ids at next edit."`
8 validation: `beans show devilGo-i56j` shows 13 epic children; decisions section intact.
9 related: all 13 epics below.
VERDICT: TRIAGED-OK


## TRIAGE WRITE-BACK ENRICHMENT (2026-10-16, user directives)
- Every bean (93/93) now carries a full `## IMPL NOTES (triage 2026-10-16)` section: all 9 triage elements (root, local probes with file:line evidence, online findings, solution, understanding, repro, fix commands, validation, related) — maximum implementation detail captured in-bean.
- Status router applied: all 26 TRIAGED-OK tasks flipped draft→todo (27 todo total incl. earlier promotions). 56 NEEDS-WORK + 1 DECISION-NEEDED stay draft until fix pass lands their corrections.


## LAYOUT DECISION (user-confirmed 2026-10-16)
AGENTS.md three-layer layout GOVERNS: internal/{engine,game,diablo}, binaries cmd/devilutionGo + cmd/devilutionGo-headless. Requirements.md Appendix E 14-package tree is SUPERSEDED — wherever a bean Scope/AC cites Appendix E paths (internal/world, entity, gameplay, render, ui, net, assets, persist, input, platform, demo, cmd/dgo), they map: world/gameplay→diablo, entity→engine/ecs, render→engine/render, ui→game/ui, net→engine/net + diablo/net, assets→engine/asset + diablo/assets, persist→engine/save + diablo/save, input→engine/input, platform→engine/config, demo→engine (headless/demo tags), cmd/dgo→cmd/devilutionGo.


## EXTERNAL LIB POLICY (2026-10-16, user question — binding)
Canonical lib catalog = docs/go-libraries.md (19 pinned core deps + must-port list). Beans must name their libs EXPLICITLY in a `## Libraries` section (module path + purpose + go.mod pin) so a worker never picks a substitute. Lib-choice changes = decision entry in milestone, never silent.


## FIX PASS COMPLETE (2026-10-16)
- 41 declared-but-unwired blocked_by edges created in tracker (parsed from triage blocks, qualified ids only).
- 8 count/premise corrections appended (TalkID 24, handlers 80, row-count false AC, quest_dialog 25/6/312, gold boundary, AddLight inventory, PerPixel/PartiallyLit, CMD_LIGHT premise rewrite).
- 8 decision-alignment notes (extract-at-build, per-hero DBs, diablo.ini, Lua deferral) + 3 overlap seams resolved (canonical pointers).
- 15 dep-remap/cite/stale-count fix notes; vinx M4 audit re-pinned to live 160 gbIsHellfire refs.
- ALL 56 NEEDS-WORK addressed → verdict TRIAGED-OK (resolved); tasks promoted: ledger now 79 todo (all implementation tasks) / 14 draft = 1 milestone + 13 epics by design (epics stay open containers).
- Ledger is DISPATCH-READY: natural dependency order, todo = claimable.
