---
# devilGo-noeb
title: Options ini import SimOptions freeze
status: completed
type: task
priority: normal
created_at: 2026-11-02T19:34:06Z
updated_at: 2026-11-02T20:15:20Z
---

## Goal

Typed options tree (internal/engine/config.go + ini_import.go) matching
upstream options.h/cpp: parse, validate, persist, migrate.

## Libraries

stdlib only. Skills: coding-rules, ponytail, go-dev.

## Triage

brainer pass owed pre-merge.

## Skills for coder

~/.agents/skills/coding-rules/SKILL.md
~/.agents/skills/ponytail/SKILL.md
~/.agents/skills/go-dev/SKILL.md

## Acceptance criteria

- [x] ini_test port green (upstream test/ini_test.cpp; 15 ini tests + 7 config tests)
- [x] TickRate maps 20..50 like the gamemenu slider (diablo.cpp:146
      gnTickDelay); audio rate restricted {22050, 44100, 48000}
- [x] Unknown/legacy ini keys migrate or are ignored 1:1 with upstream
      behavior (no error spam)

## Progress log

placeholder

## Progress log

- 2026-11-02: lane salvaged by parent after two worker stalls (rc=3). Commits on dg-options-ini-import-simoptions-freeze: 921dbc0 fixtures, 0d72d2f ini parser/serializer port, 2e9c4fd SimOptions typed tree, 0736a75 byte-identical round-trip test. Full gates green.

- 2026-11-02: brainer FIX round actioned (02ab54e): F1 comment-span interior blanks (ini.cpp:151-160), F2 tab value trim (ini.cpp:206), F3 exact space/tab trims for key + section-after, F4 config.go citation corrected + deviation recorded + TickRate clamp, F5 prevR comment normalization (ini.cpp:389), F6 error-branch tests, F7 Changed/MarkAsUnchanged accessors (options.cpp:166/180). Gates re-run green.
- 2026-11-02: CI green on final head db13f2c (run 41517104453; also 41517052386 @ 02ab54e, PoW-fix run @ 2dfba7b). Bean closeout.
## Proof of Work

- Fixtures byte-verbatim from test/ini_test.cpp input/expected (internal/testdata/config/).
- ini_import.go mirrors Ini::parse/serialize incl. comment spans, BOM skip, key-not-in-section error (ini.cpp:137-215, 415).
- SimOptions defaults: TickRate 20 (options.cpp:843), SampleRate 22050 {22050,44100,48000} (options.cpp:55,520); gnTickDelay=1000/tickRate (gamemenu.cpp:277).
- Gates: go test ./... 42 ok; scripts/lint.sh ./... 0 issues at 0736a75.
