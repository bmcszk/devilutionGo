---
# devilGo-nxxo
title: 'Gameplay: combat, spells/missiles, items, inventory, stores, quests, XP'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-2
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-15T16:37:34Z
parent: devilGo-i56j
blocked_by:
    - devilGo-97mn
---

# Gameplay systems

## Goal
Everything that makes it a game: to-hit/damage, 50+ spells via missile table, item generation with affixes/uniques, inventory/belt grid, store transactions, quest machinery, XP/levels, spawn gating, Hellfire branches.

## Research
- Combat: player.h:575-666 (to-hit formulas, armor pierce); player.cpp:524-660 PlrHitMonst full pipeline (roll order, <<6 fixed point, class crit, undead/mace, Jesters/Devastation/Doppelganger hellfire); monster→player :712.
- Spells/missiles: tables/spelldat.h:225-252; tables/misdat.h:166-196 fn pointers → registry; missiles.cpp:479 CheckMissileCol, :636 Move…, AddMissile chains :728-1242; misdat.tsv + missile_sprites.tsv.
- Items: tables/itemdat.h:628-655 PLStruct/UniqueItem; items.cpp:1048 SaveItemAffix, :1138/:1218-1224 affix selection, :3195 SpawnUnique, :3422 SpawnItem; durability break player.cpp:505-522.
- Inventory: player.h:38-39 (40 grid, 8 belt); inv.cpp placement; tests inv_test, inventory_ui_test, items_test.
- Stores: stores.h:44-70 TalkID; stores.cpp state machine + tests stores/store_transaction/vendor/visual_store.
- Quests: tables/questdat.hpp:102-129; quests.cpp:60-110 InitQuests (spawn wipe :96, pools from DungeonSeeds[15]), :227 CheckQuestKill, :397 ResyncQuests; MAXQUESTS 24.
- XP: txtdata/Experience.tsv via tables/playerdat.cpp:91; class data :203 baseMeleeToHit, :379 GetClassAttributes.
- Spawn gating: monster.cpp:3159, player.cpp:3479,3688-3701; hellfire branch count 135 root-cpp sites.
- Study: /tmp/dg-requirements.md §4, Appendix J.1.

## Scope
internal/gameplay + stores subpackage; combat pipelines; missile registry; itemgen; inventory ops; quest FSM; XP; SimOptions gating helpers gbIsSpawn/gbIsHellfire.

## Acceptance criteria
- [ ] Combat parity: 1000-seed roll-tapes for PlrHitMonst (warrior/rogue/sorcerer vs 5 monster classes) equal C++ dumps
- [ ] Every missile in misdat.tsv has Add+Process registered; missiles_test.cpp port passes
- [ ] Itemgen: affix tables load; 10000-seed generation distribution matches C++ histograms (chi-square p>0.01) for 3 item classes
- [ ] Inventory: all inv_test cases; belt usage; gold stacking
- [ ] Stores: store_transaction_test port passes; full Smith/Witch/Healer/Boy flows scripted-test green
- [ ] Quests: quests_test port; 16 Diablo quests lifecycle scripted; spawn wipe verified
- [ ] XP: level thresholds table equal; level-up stat point flow
- [ ] Hellfire: damage branches (Spectral Arrow fix, Jesters) behind gbIsHellfire flag with tests on both settings

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — function-registry pattern replacing fn pointers
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded tape tests
/home/blaze/.agents/skills/coding-rules/SKILL.md — combat pipeline reads like the C++ steps

## Dependencies & blockers
blocked_by dg-epic-04. Blocks dg-epic-08 (save serializes gameplay state), milestones M2+.

## Open questions
- H3 affix representation (codegen vs hand) resolved in dg-task-008.
