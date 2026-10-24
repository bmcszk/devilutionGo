# Requirements volumes — layout + regeneration recipe

`docs/requirements.md` (committed to devilutionGo) is a concatenation, in order, of:

1. **Volume I — architecture** (12 subsystem sections: core loop/threading, world/levels, entities, gameplay, persistence, asset pipeline, rendering, networking, input, UI flow, platform/misc, testing). Each: purpose, key sources with `file:line`, data structures, algorithms, determinism constraints, Go/Ebiten mapping, risks.
2. **Volume II — data inventories** V0–V12. V0.1 TSV format contract; V0.2 Hellfire `mods/hf/txtdata` override+append mechanism; V1 items (436 rows: 168 bases + 83 prefixes + 95 suffixes + 90 uniques, affix algorithm + drop rules); V2 monsters (212 rows + AiProc table 40 entries); V3 missiles (114 rows + Add*/Process* registry ~74 fns); V4 objects (109 rows + shrines); V5 town/NPCs/stores/cow content; V6 quests (24 + state machine); V7 spells (36 + formulas); V8 text/menus (477 rows + every menu tree); V9 audio (829 effects rows + music + movies); V10 classes+XP (1080 + 50 rows); V11 levels (trigs/themes/setmaps); V12 misc enums (cursors/gold/durability/gating/QoL).
3. **Volume III — per-file sweep**: 56 batches × ≤12 files over all 670 `Source/`+`test/` files; one `### <path>` entry each with Purpose / Public API / Depends on / Go-port requirements / 1:1 constraints (real line-cited constants) / determinism / test coverage / Go mapping / risks. 668/670 covered (2 trivial DiabloUI headers).
4. **Graphics G1–G11**: frame architecture (640×480 fixed, FilterNearest), tile micro-render pipeline, CEL/CL2/CLX/PCX byte formats with worked decode examples, palette/TRN + light tables, entity/UI sprite plumbing, fonts, effects, asset manifest, consolidated Ebiten mapping decisions (index framebuffer + 16 palette-row textures + Kage palette-lookup shader; atlas ≤4096²; batching = same atlas+blend+filter runs; Update/Draw split with tick-progress interpolation), pixel-parity test plan (decoder goldens, C++ backbuffer dumper vs Go, seed-dungeon CI goldens, GPU-variance rule).
5. **Sound S1–S9**: architecture map sound.cpp/soundsample/effects.cpp → Go packages; exact math (VolumeScale 3321.9281, millibel remap, StereoSeparation 6000, pan/volume ±6400, VOLUME_MIN -1600, ATTENUATION_MIN -6400, 80 ms retrigger dedup, DuplicateSound pool); streaming/music loops (InfiniteLoop, full vs spawn track tables, mp3 twin resolution); effects catalog; audio options (context once-per-process — decide the re-init story); Ebiten mapping (port the StereoPanStream L/R-gain pattern; mono→stereo upsample); parity tests (math golden tables, dedup boundary with fake clock, catalog row counts).

## Regeneration recipe (any volume lost / extended)

- Ground truth for row counts: `wc -l` on `assets/txtdata/**/*.tsv` minus header rows; per-file sweep inventory: `find Source test -name '*.cpp' -o -name '*.h' -o -name '*.hpp' | sort`.
- Dispatch via pi RPC (see `dev/pi-delegation` references/rpc-mode.md): contract = role + GIT JAIL + single /tmp output per lane + completion marker + self-run VERIFY block (line floors, section-heading presence, distinct-cited-path count, JSON validity, skills paths exist).
- Batch fanout for sweeps; resumable driver; gate on the output FILES (split variants summed), never the lane's chat summary.
- After the run: validate structurally in code, import via `scripts/import_bean_tree.py`, retrofit the three program mandates, wire blockers through the plan-id map.
- Network note: simultaneous all-lane failure with `curl` 000 responses while ICMP passes = egress outage — wait and re-invoke, don't rewrite the contract.

## Bean-id plan namespaces (avoid collisions)

- `dg-task-0NN` Volume-I tasks · `dg-task-1NN` Phase-7 data tasks · `dg-task-2NN` graphics extras · `dg-task-3NN` sound tasks · `dg-task-4NN` gap-area tasks. Real bean ids live in the repo `.beans/`; keep the plan-id→bean-id map (json) alongside the ledger commit so later waves can wire `--blocked-by` across namespaces.