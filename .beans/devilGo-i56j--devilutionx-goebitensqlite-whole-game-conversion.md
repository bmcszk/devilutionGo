---
# devilGo-i56j
title: devilutionX → Go+Ebiten+SQLite whole-game conversion
status: todo
type: epic
priority: normal
created_at: 2026-10-15T15:59:25Z
updated_at: 2026-10-29T16:05:22Z
---

## ROADMAP (2026-10-29 comprehensive review — validated against repo + devilutionX pin c20b2f245)

### Where we are (validated)
PR #1-#3 merged: engine core (RNG/path/crawl/tick), txtdata TSV stack (jx08/zvwd + 10 gamedata
packages), asset decoders (CLX/CL2/CEL/PCX/TRN/palettes + goldens), missiles (process engine +
misdat fn-name registry, ~10 behaviors), combat (PlrHitMonstCore pipeline), entities structs,
save codec+diablosha, demo format+record/replay, light tables. Suite green: 27 packages, ~10.9k
test cases. NOT yet: rendering (0%), DRLG (0%), UI (0%), networking (0%), SQLite store (0%).

### Wave sequencing (dependency-ordered)
- W1 renderer core: gk5e tiles -> ru6l iso -> 66o7 framebuffer -> m18k light (M1 gate)
- W1 parallel: 5wsb finishers l5fp + r7qf + ow6p (unblocks world/audio)
- W2 world: yeia town -> 76tu L1 -> v791 monsters -> foug themes (M1 gate)
- W2 parallel: gameplay chain mwlm(finish) -> 2j8j -> baxb -> onlm -> rfrj
- W3 M1 integration: dcob+emdz hero/menu, c6wt input, mlv2 panels, oqho+xwtt save, 41p0+uc44 gates
- W4 M2 push: 4ybb L2-L4 DRLG, xw8k quests FSM, 0qnd stores, 9o2h vision, doom/dead/gamemenu/appfat fillers
- W5 M3 networking: zo8j lane (vgxq done: packet codec) -> TCP host/join -> delta sync -> trade
- W6 M4 hellfire: hf TSV stack (fixtures exist), Monk, crypt/nest, 8 quests, gbIsHellfire audit

### Known gaps (created as beans or tracked here)
- mwlm honest state: 0/50 Add fns implemented, registries unpopulated (only name lists);
  ~10/36 Process behaviors done. Finish before items chain claims it as done.
- Small cpp files with NO beans: dead.cpp corpses (108L), doom.cpp debug (44L), appfat.cpp
  (106L), gamemenu.cpp (391L), capture.cpp SMK (deferred per GI-5), console.cpp (668L, M2+).
- 49 completed beans carry unticked AC checkboxes (PoW-style closure); accepted: work verified
  on master + CI green. New rule: tick ACs at close time going forward.

### Milestones
M0 done. M1 devilGo-ftbu (35 open beans across 8 features). M2 devilGo-vinx.
M3/M4 later waves. Support lane: 1n0a (gates), klu5/ocl1 quality sweeps.

### TREE STRUCTURE DECISION (2026-10-29, 'epics under milestones?' resolved)
- GOAL epic (this bean) spans M1-M4 -> cannot sit under ONE milestone; stays ROOT.
- Milestones are ROOT (beans: milestone has no legal parent). They GATE via integration
  manifests in their bodies (ftbu/vinx), they do not contain work.
- Legal parent edges: epic<-{milestone}, feature<-{milestone,epic}, task/bug<-{milestone,epic,feature}.
- Features are milestone-CROSSING (Renderer serves M1-M4) -> they stay under the GOAL epic.
  Putting them under M1 would orphan M2-M4 tracking.
- SUPPORT epic devilGo-p1t8 created (user request 'support epic for lints, tests, tools'):
  1n0a (testing&parity) + completed quality waves (qna9, lu2v, gmtn, kefy, qo09, gt3f)
  moved under it. Game features stay under GOAL.
- Result: 4 roots (M1, M2-4, GOAL epic, SUPPORT epic). Check green.
