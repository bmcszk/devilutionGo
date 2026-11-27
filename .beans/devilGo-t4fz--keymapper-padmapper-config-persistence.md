---
# devilGo-t4fz
title: Keymapper + padmapper + config persistence
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-26T22:15:32Z
parent: devilGo-k736
---

# Keymapper + padmapper + config persistence

## Goal
The keymapper BIND MODEL in internal/engine: Key keycode namespace (incl. KeymapperMouseButtonMask + scroll codes), verbatim key-name tables, Action/Keymapper with AddAction(front)+CommitActions(reverse), SetValue duplicate-unbind, LoadFromIni/SaveToIni on section Keymapping (absent->default, empty->unbound, unknown->default), Press/Release with the upstream guard order + text/number-entry carve-outs (injected flags struct, never globals). Defaults measured into a committed TSV fixture (corpus-measurement law). THE DATA HALF ONLY - this wave.

## Research
- Model: KeymapperOptions ctor tables (options.cpp:1108-1178); Action LoadFromIni :1210-1234, SaveToIni :1236-1249; SetValue :1262-1285; AddAction :1291 pushes front, CommitActions :1296 reverses. Section Keymapping :1110.
- Dispatch: keymapper.cpp:53-77 guard order (unmapped / nil handler / !isEnabled / ChatFlag), ToAsciiUpper :37-49, release carve-outs IsTextEntryKey :22-30, IsNumberEntryKey :32-35.
- Defaults at CALL SITES: InitKeymapActions (diablo.cpp:1839-2157), ~100 entries incl. loop-generated BeltItem1..8, QuickSpell1..12, QuickMessage1..16; NOEXIT + _DEBUG conditionals EXCLUDED (release build). Row count MEASURED by the coder from git show e1ad0c2b6:Source/diablo.cpp - never a literal.
- Mouse/scroll synthetic keys: mask 1<<31, scroll codes 65536..65539 (options.h:711-715).
- Padmapper DEFERRED: zero Go gamepad code exists; home = new draft bean (filed at wave close).
- Callbacks DEFERRED: the ~100 actions bind gameplay verbs that do not exist in Go yet; home = new draft bean (filed at wave close).

## Acceptance criteria
- [x] AC2 (keymapper model): engine.Keymapper loads/saves [Keymapping] 1:1 with X:options.cpp KeymapperOptions::Action (absent->default, empty->unbound, unknown->default, duplicate rebinds and unbinds the previous owner), key-name tables round-trip, Press/Release apply the X:keymapper.cpp guard order; defaults equal the committed keymapping_defaults.tsv (row count asserted from the fixture, measured from git show e1ad0c2b6:Source/diablo.cpp). RED: go test ./internal/engine -run TestKeymapper -count=1, behavioural RED for the defaults table pasted verbatim.

## Libraries
stdlib only (internal/engine stays stdlib-only).

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/port-parity-testing/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Combined triage 2026-11-26 (task-opts-triage-brainer): evidence E4 + scope split + file tables in .hermes/lane/options-slice/triage.md. Padmapper + callbacks deferred to named drafts under the Input epic.



## Proof of Work (Options-slice wave, parent-verified 2026-11-26)

- AC2 landed: internal/engine/keymapper.go (tables verbatim at e1ad0c2b6, reviewer spot-checked 5 entries incl. KEYPADNUM/LALT/X1MOUSE/scroll codes), duplicate unbind pinned, guard order line-for-line vs keymapper.cpp:53-77, keymapping_defaults.tsv committed with the row count asserted from the fixture (58 named + loop-generated; NOEXIT + _DEBUG excluded; QuickMessages corrected to 10 by measurement).
- RED: the empty-table walk (bound nothing -> every fixture row differs) pasted verbatim in the wave log; GREEN internal/engine ok.
- Split recorded: padmapper -> devilGo-fxui; callbacks + routing -> devilGo-hi7t (both filed, Input epic).
- Status: AC2 ticked from parent re-run evidence (TestKeymapper battery ok, mutation live-verified in S3's volume battery + the TSV mutation in the wave log).
