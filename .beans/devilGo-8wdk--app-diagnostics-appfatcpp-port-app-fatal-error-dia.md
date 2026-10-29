---
# devilGo-8wdk
title: 'App diagnostics: appfat.cpp port (app_fatal, error dialogs)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-10-29T16:14:42Z
parent: devilGo-yjuv
---

## Goal

Port `appfat.cpp` (106 lines): application-fatal error handling — `app_fatal` +
`ErrSdl`/`ErrSdlDlg`, `DiabloError` dialog plumbing, `stderr` logging path.

## Research

Source/appfat.cpp:1-106; appfat.h; diablo.cpp `FatalError` call sites;
SDL_ShowSimpleMessageBox semantics vs Ebiten (use ebiten-less stdlib dialog or log-only
in headless build tag; GUI path only in the windowed cmd target).

## Scope

internal/platform/appfat.go (or internal/engine if platform imports are wrong per AGENTS
layer rules — check first). Headless: log+exit(1). Windowed: message box.

## Acceptance criteria

- [ ] All appfat.cpp call sites have Go equivalents (grep parity: app_fatal/ErrSdl)
- [ ] Headless mode exits non-zero with the C++-matching message on stderr
- [ ] Windowed mode shows an error dialog (manual check documented in bean)
- [ ] No new imports violating the AGENTS.md layer rules

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md

## Dependencies & blockers

None. Small, self-contained. Any wave.
