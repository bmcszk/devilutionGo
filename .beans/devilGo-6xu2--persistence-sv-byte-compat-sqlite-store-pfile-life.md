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
updated_at: 2026-10-16T14:01:03Z
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
