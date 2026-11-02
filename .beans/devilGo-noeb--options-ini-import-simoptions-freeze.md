---
# devilGo-noeb
title: Options ini import SimOptions freeze
status: in-progress
type: task
priority: normal
created_at: 2026-11-02T19:34:06Z
updated_at: 2026-11-02T19:34:31Z
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

- [ ] ini_test port green (upstream tests/ini_test.cpp)
- [ ] TickRate maps 20..50 like the gamemenu slider (diablo.cpp:146
      gnTickDelay); audio rate restricted {22050, 44100, 48000}
- [ ] Unknown/legacy ini keys migrate or are ignored 1:1 with upstream
      behavior (no error spam)

## Progress log

placeholder

## Progress log

- 2026-11-02: lane salvaged by parent after two worker stalls (rc=3). Commits on dg-options-ini-import-simoptions-freeze: 921dbc0 fixtures, 0d72d2f ini parser/serializer port, 2e9c4fd SimOptions typed tree, 0736a75 byte-identical round-trip test. Full gates green.

## Proof of Work

- Fixtures byte-verbatim from test/ini_test.cpp input/expected (internal/testdata/config/).
- ini_import.go mirrors Ini::parse/serialize incl. comment spans, BOM skip, key-not-in-section error (ini.cpp:137-215, 415).
- SimOptions defaults: TickRate 20 (options.cpp:843), SampleRate 22050 {22050,44100,48000} (options.cpp:55,520); gnTickDelay=1000/tickRate (gamemenu.cpp:277).
- Gates: go test ./... 42 ok; scripts/lint.sh ./... 0 issues at 0736a75.
