---
# devilGo-yjuv
title: 'Platform: options, i18n, Lua dev tooling, QoL modules'
status: todo
type: feature
priority: normal
tags:
    - plan:epic
    - phase-3
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-29T16:13:18Z
parent: devilGo-i56j
---

# Platform layer

## Goal
Options system (+ini migration), translations (.po/.mo), gopher-lua dev REPL subset, QoL: stash, xpbar, item labels, floating numbers, monhealthbar, autopickup, chatlog.

## Research
options.h:877, :512-530 + options.cpp:55,520 (rates); ini utils + ini_test; language.cpp:39-42 .mo/.gmo; 25 .po in Translations/; lua modules (audio.cpp:37-47 playSfx surface; dev/* debug modules; repl+autocomplete; GameStart hook diablo.cpp:895; integrity gate :900-903); qol/ 16 files (stash 100 pages loadsave.cpp:2740-2760; xpbar; itemlabels; floatingnumbers; monhealthbar; autopickup; chatlog; visual_store). Study §11.

## Scope
internal/platform/{options,i18n,lua,qol}. SimOptions snapshot freeze at game start.

## Acceptance criteria
- [ ] Options load/save JSON + diablo.ini import (ini_test port)
- [ ] i18n: en+pl catalogs load, LanguageTranslate parity for 100 keys; UI strings translated end-to-end
- [ ] Lua: gopher-lua boots assets/lua/inspect.lua; playSfx binding works; REPL over stdin in headless build
- [ ] QoL feature-flagged modules: stash (grid ops + save), xpbar (curve from Experience.tsv), item labels, floating damage numbers, monster healthbar
- [ ] Sim freeze: changing graphics options mid-game cannot alter logic (test asserts SimOptions immutable snapshot)

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
/home/blaze/.agents/skills/go-dev/SKILL.md — registry/config patterns

## Dependencies & blockers
blocked_by dg-epic-05 (numbers/economy for labels), dg-epic-08 (stash persistence). M3+ for chatlog.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DEFERRED — question answered by recorded milestone decision; revisit point documented in the question text above.
- H15 Lua scope: dev tooling only (decided); revisit only if mods milestone.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:options AC says JSON (contradicts diablo.ini-1:1 amendment) + Lua ACs not deferred per DECIDED #3


## IMPL NOTES (triage 2026-10-16)
1 root: Options/i18n/Lua/QoL platform layer. Wired blockedBy nxxo ✓ (body "dg-epic-05"). 5 children. TWO decision conflicts in body.
2 local: Cites verified: options.h:877 `struct Options` ✓; utils/language.cpp:39-42 .mo/.gmo ✓ (bean says "language.cpp" — path prefix only, content exact). BUT (a) AC "Options load/save JSON + diablo.ini import" contradicts SETTINGS FORMAT AMENDMENT (71b17dd): diablo.ini INI is the 1:1 read/write format (utils/ini.cpp semantics), NOT JSON; (b) Goal+AC carry gopher-lua REPL/playSfx as in-epic deliverables while DECIDED #3 defers Lua post-1.0 (engine keeps 0-VM LuaEvent stubs).
3 online: n/a.
4 solution: (a) Replace JSON AC with INI codec AC (sections/keys/comments preserved, byte-copyable between devilutionX and Go). (b) Split Lua out: in-epic scope = LuaEvent hook stubs (0 VM) only; gopher-lua REPL AC moves to a deferred post-1.0 bean (decision: NOT cut). Keep QoL + i18n + SimOptions-freeze ACs as-is (all testable).
5 understanding: engine/config + platform layers; SimOptions freeze at game start is the determinism guard for options — keep.
6 repro: `sed -n '877p' Source/options.h`; `grep -n "SETTINGS FORMAT AMENDMENT\|diablo.ini" .beans/devilGo-i56j--*.md | head -5`; `grep -n "ini.cpp" Source/utils/ini.cpp | head -1`
7 fix: `beans update devilGo-yjuv --body-replace-old "Options load/save JSON + diablo.ini import (ini_test port)" --body-replace-new "Options load/save 1:1 diablo.ini INI (utils/ini.cpp semantics: sections/keys/comments preserved, byte-compatible with devilutionX; ini_test port) — no JSON settings"` ; `beans update devilGo-yjuv --body-replace-old "Lua: gopher-lua boots assets/lua/inspect.lua; playSfx binding works; REPL over stdin in headless build" --body-replace-new "Lua DEFERRED post-1.0 (DECIDED #3): this epic ships LuaEvent hook-point stubs (0 VM) only; gopher-lua REPL/playSfx ACs moved to deferred post-1.0 feature bean"`
8 validation: body INI-first + Lua-deferred; children 0tsf (ini import) and yd48 (gopher-lua tooling) checked for consistency in batch-09.
9 related: blocker nxxo; children 0tsf, sr71, yd48, cciu, orkj.
VERDICT: NEEDS-WORK:options AC says JSON (contradicts diablo.ini-1:1 amendment) + Lua ACs not deferred per DECIDED #3

Decision-alignment (triage): settings = diablo.ini 1:1 (INI codec port), NOT JSON. Lua ACs marked deferred post-1.0 with gopher-lua (pure-Go).

## TRIAGE 2026-10-18 (pi wave, batch 3)

devilGo-yjuv** (Platform epic, parent i56j)
1 root: options/i18n/lua/qol; gate nxxo ✓ wired; **two unresolved body conflicts** from prior NEEDS-WORK.
2 local: `options.h:877 struct Options` ✓; `utils/language.cpp:39-42` .mo/.gmo ✓ (25 .po files ✓); stash 100 pages `loadsave.cpp:~2743` ✓; options.cpp sampleRate :520 ✓. 5 children exist (0tsf,sr71,yd48,cciu,orkj).
3 online: SKIPPED.
4 solution: prior fix was **notes-only stamp** ("Decision-alignment (triage)" line) — AC text unedited: (a) AC "Options load/save **JSON** + diablo.ini import" still contradicts INI-1:1 (child 0tsf already INI-correct ✓); (b) Lua AC "gopher-lua boots…REPL over stdin" still in-body while DECIDED #3 defers Lua post-1.0 (child yd48 split ✓). Also same dep gap as zo8j: prose dg-epic-08→6xu2 (stash persistence) unwired in frontmatter.
5 DDD fit: engine/config + platform/{options,i18n,qol} ✓.
6 dep-order: nxxo + 6xu2 → yjuv; chatlog M3+.
7 AC-gap: replace JSON AC with INI codec AC (ini.cpp semantics, byte-copyable); rewrite Lua AC to LuaEvent stubs (0 VM) in-epic, gopher-lua deferred; wire 6xu2.
8 validation: ini_test port green; SimOptions freeze immutability test; en+pl 100-key parity.
9 related: nxxo, 6xu2, 0tsf, sr71, yd48, cciu, orkj.

**NEEDS-WORK:AC still says JSON (INI-1:1 amendment) + Lua AC not deferred per DECIDED #3 — prior fix was note-only, body unedited; + 6xu2 dep unwired**


Wave outputs: /tmp/dg-triage/out-b3.md

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
