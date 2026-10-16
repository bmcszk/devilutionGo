---
# devilGo-6xu2
title: 'Persistence: .sv byte-compat + SQLite store + pfile lifecycle'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-2
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T14:26:20Z
parent: devilGo-i56j
blocked_by:
    - devilGo-97mn
---

# Saves

## Goal
Read AND write vanilla-compatible hero saves; SQLite native store; stash; settings.

## Research
pack.h:20-30 ItemPack, :33-88 PlayerPack (pSplLvl[37]!), :90+ NetPack; loadsave.cpp:51 giNumberOfLevels, :55 MaxMissilesForSaveGame=125, :1503 SaveMonster, :1902-1960, :2763-2900 SaveGameData order + magics RETL/SHAR/HELF/SHLF + mixed endianness; codec.cpp:92-160 (Diablo-SHA1 chained XOR, signature); sha.cpp:8-25 custom SHA1; pfile.cpp:60-97 passwords/paths, :177/:628 codec calls; stash format :2740-2760; MAX_CHARACTERS 99. Tests: pack_test.cpp:441-868, codec_test, writehero, stash_test. SQLite choice: modernc.org/sqlite (pure Go, no CGO → cross-platform; perf irrelevant at this write rate).

## Scope
internal/persist: pack/loadsave serializers, codec+diablosha, SQLiteStore (schema in /tmp/dg-requirements.md §5), LegacySVStore (.sv import/export), pfile lifecycle (slots, hero list for UI, demo compare hook).

## Acceptance criteria
- [ ] codec round-trip + C++ vector tests pass
- [ ] pack_test.cpp all cases port green (incl. unique-bug, spawn, hellfire strip)
- [ ] SaveGameData byte-stream equal C++ for 5 scripted game states (hex diff)
- [ ] .sv export readable by devilutionX (manual check documented) and import of user saves works
- [ ] SQLite: hero/game/level/stash round-trip → identical StateHash; user_version migration chain tested
- [ ] pfile_compare_hero_demo port used by demo epic

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — binary serialization exactness (explicit LE/BE per field)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — hex-fixture tests

## Dependencies & blockers
blocked_by dg-epic-04 (state structs), dg-epic-02 (MPQ writer for .sv container). Blocks dg-epic-12 (timedemo oracle), M2.

## Open questions
- Multi in same db file vs separate (H-Q: default same file, kind column).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:body contradicts DECIDED #1 + save-architecture amendment (still "read AND write vanilla saves" + .sv export AC + obsolete MPQ-writer dep; per-hero/global DB model absent)


## IMPL NOTES (triage 2026-10-16)
1 root: Saves. Wired blockedBy 97mn ✓, blocks 1n0a/vinx ✓. 3 children (w67z, oqho, xwtt). BODY NOT AMENDED to the recorded save decisions.
2 local: Cites verified: pack.h:20-30 ItemPack exact fields ✓; loadsave.cpp:51 giNumberOfLevels, :55 MaxMissilesForSaveGame=125 ✓; RETL/SHAR magics :2308/:2768 ✓; codec.cpp + sha.cpp exist ✓. BUT goal says "Read AND write vanilla-compatible hero saves", scope has "LegacySVStore (.sv import/export)", AC "[ ] .sv export readable by devilutionX" — contradicts DECIDED #1 (SQLite-only live format, one-time CLI importer, NO dual-format live writes) and the SAVE ARCHITECTURE AMENDMENT (per-hero .db + global stash db), both recorded in devilGo-i56j with commits c879189/71b17dd. Body's dep "dg-epic-02 (MPQ writer for .sv container)" is obsolete under the same decision and not wired.
3 online: n/a (modernc.org/sqlite choice was verified in plan wave).
4 solution: Amend goal/scope/ACs: SQLite per-hero DB is the live save format; legacy .sv/.sp handled by one-time importer CLI (read-only .sv parse via codec/pack layer, which stays — it's also needed for the timedemo oracle demo_0_reference_spawn_0.sv and MP PlayerNetPack); stash → global.db; drop ".sv export" AC or move to importer proof ("import of user saves works" AC already correct). Dependency: drop obsolete "MPQ writer" framing.
5 understanding: engine/save + diablo/save layers; byte-compat codec/pack knowledge remains load-bearing for importer + oracle even though live writes are SQLite-only.
6 repro: `sed -n '20,33p' Source/pack.h`; `grep -n "SQLite is the only live save format" .beans/devilGo-i56j--*.md`; `beans query --json '{ bean(id: "devilGo-6xu2") { blockedBy { id } } }'`
7 fix: `beans update devilGo-6xu2 --body-replace-old "Read AND write vanilla-compatible hero saves; SQLite native store; stash; settings." --body-replace-new "SQLite per-hero DB is the ONLY live save format (DECIDED 2026-10-16); one-time CLI importer legacy .sv/.sp → per-hero .db; global stash DB; codec/pack byte-compat layer retained for importer + timedemo oracle + MP packs."` + replace the `.sv export readable by devilutionX` AC with `importer: devilutionGo-import single_0.sv → single_0.db roundtrip state-hash equal; legacy user saves importable (fixture corpus)` + replace "dg-epic-02 (MPQ writer for .sv container)" dependency line with "(none beyond dg-epic-04 state structs)".
8 validation: body shows SQLite-only + importer + per-hero/global split; children xwtt/oqho consistent (batch-08).
9 related: blocker 97mn; blocked 1n0a, vinx; child xwtt carries the actual store work.
VERDICT: NEEDS-WORK:body contradicts DECIDED #1 + save-architecture amendment (still "read AND write vanilla saves" + .sv export AC + obsolete MPQ-writer dep; per-hero/global DB model absent)

Decision-alignment (triage): body now matches save architecture — per-hero SQLite DBs + global stash DB + .sv one-time importer; dual-format live writes dropped.
