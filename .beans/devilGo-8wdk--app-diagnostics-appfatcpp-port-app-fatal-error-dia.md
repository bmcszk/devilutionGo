---
# devilGo-8wdk
title: 'App diagnostics: appfat.cpp port (app_fatal, error dialogs)'
status: in-progress
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-03T00:54:32Z
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
- [x] Headless mode exits non-zero with the C++-matching message on stderr — TestExitOnError_Fatal_ExitsOneWithStderrMessage (cmd/devilutionGo-headless/main_test.go:20): exit code 1 and stderr "Error\ntest\n" (appfat.cpp:66 diablo_quit(1))
- [ ] Windowed mode shows an error dialog (manual check documented in bean)
- [x] No new imports violating the AGENTS.md layer rules — internal/engine/appfat imports stdlib only (fmt, io, strings); internal/diablo/appdiag imports engine/appfat + fmt. Verified by reading both files' import blocks; go vet ./... rc0.

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md

## Dependencies & blockers

None. Small, self-contained. Any wave.

## Libraries

Binds ebiten v2 for rendering (owner flag in Scope); everything else stdlib.
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).

## Deviations accepted (brainer PR #5 review)

- Windowed dismiss-bit AC stays OPEN: gated on the UI wave (Ebiten not in go.mod). PR #5 explicitly defers it — headless parity done.
- Exit code: headless exits 1 (appfat.cpp:66 diablo_quit(1)); dtogo window path uses 64 (diablo.cpp:1093) — ported at the cmd boundary only.
- ErrSdl deferral recorded: low-level SDL surface lands with the renderer bean.

## Progress log
- 2026-10-29: claimed in-progress, coder dispatched (session devilGo-8wdk-coder)
- 2026-10-31: PR #5 open (dg-app-diagnostics-appfatcpp-port), CI green, brainer verdict SHIP; windowed-dialog AC deferred to UI wave (see Deviations accepted)


## Findings (2026-11-03, asset verification wave — read-only gate pass)

Gates re-run here: `go build ./...` rc0, `gofmt -l internal/ cmd/` empty,
`go vet ./...` rc0, `scripts/lint.sh ./internal/... ./cmd/...` 0 issues,
`go test ./cmd/... ./internal/diablo/appdiag/... ./internal/engine/appfat/...` ok.

- AC "All appfat.cpp call sites have Go equivalents" is NOT met and cannot be
  met yet: upstream has 111 `app_fatal(` call sites plus 80 `ErrDlg(`/`ErrSdl`
  sites in Source/, and the Go callers (renderer, UI, loadsave) are not ported.
  The helper surface is complete for what exists: AppFatal, AssertionFailed,
  InsertCDDlg, DirErrorDlg, ErrDlg (internal/engine/appfat/appfat.go:55-90)
  against upstream `app_fatal`, `assert_fail`, `InsertCDDlg`, `DirErrorDlg`,
  `ErrDlg` (Source/appfat.cpp:69-98); `FreeDlg`/`DisplayFatalErrorAndExit` are
  subsumed by FatalError.WriteTo + the cmd exit path (main.go:76).
- AC "Windowed mode shows an error dialog" stays open by the brainer-accepted
  deviation above (Ebiten absent from go.mod; UI wave owns it).
- AC2/AC4 are now ticked with the evidence named in the criteria.

Bean stays in-progress: AC1 is caller-blocked, AC3 is wave-gated.
