---
# devilGo-yjuv
title: 'Platform: options, i18n, Lua dev tooling, QoL modules'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-3
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-15T16:37:34Z
parent: devilGo-i56j
blocked_by:
    - devilGo-nxxo
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
/home/blaze/.agents/skills/go-dev/SKILL.md — registry/config patterns

## Dependencies & blockers
blocked_by dg-epic-05 (numbers/economy for labels), dg-epic-08 (stash persistence). M3+ for chatlog.

## Open questions
- H15 Lua scope: dev tooling only (decided); revisit only if mods milestone.
