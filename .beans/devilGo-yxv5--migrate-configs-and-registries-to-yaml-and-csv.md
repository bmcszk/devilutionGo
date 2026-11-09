---
# devilGo-yxv5
title: 'Migrate settings: diablo.ini checks plus registry YAML views'
status: completed
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-09T22:46:23Z
parent: devilGo-o8e7
---

# Migrate settings: diablo.ini checks plus registry YAML views

## Goal

ADR-014 settings subcommand (ADR-004 law): `devilutionGo-migrate settings`
reads diablo.ini, checks validity, carries the file over byte-identical into
the engine settings path, and emits flat YAML registry views for engine
tooling only. INI stays canonical; YAML is a view, never a second source of
truth. CSV table exports belong to the assets subcommand, not this lane.

## Research

- internal/engine/ini_import.go already reads diablo.ini and
  internal/engine/roundtrip_test.go proves byte-identical round trips.
- [ADR-004](../docs/adrs/adr-004.md): settings live in INI, not SQLite.
- [ADR-013](../docs/adrs/adr-013.md): YAML for flat registry views.

## Scope

- settings subcommand: INI validity checks, byte-identical carryover,
  flat YAML registry view writer.
- Tests: round trip on synthetic INI fixtures committed as text/goldens.

## Acceptance criteria

- [ ] Given a diablo.ini, migration carries it over byte-identical and reports violations it cannot fix.
- [ ] YAML views regenerate byte-identical on a second run.
- [x] go test ./... green, gofmt clean.

## Libraries

- None new (INI reader exists; YAML writing via the existing stdlib stack).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent accepted; runs after the engine INI round-trip plumbing it reuses.

## Open questions

- YAML split: one file per registry family unless tooling prefers one file.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.


## Summary of Changes

`settings` subcommand landed in `cmd/devilutionGo-migrate/settings.go`
(wired in `main.go`; pending-lane test removed from `main_test.go`, coverage
replaced by `settings_test.go`):

- Reuses `internal/engine.ParseIni`/`Get` (read-only; round-trip proven by
  engine tests). Typed checks via engine seams: `Game.Speed` inside
  `TickRateMin..TickRateMax`, `Audio.Sample Rate` via `ValidAudioSampleRate`;
  keys outside typed options are listed as unfixable violations, never
  rewritten (AC 1, TestSettingsReportListsViolationsItCannotFix).
- Byte-identical carryover: cmp-equal copy into the engine settings path
  (ADR-004: working dir holding portable.ini, else SDL pref dir
  `~/.local/share/diasurgical/devilution`, XDG_DATA_HOME honored);
  verified by re-read + bytes.Equal and TestSettingsCarriesIniByteIdentical.
- Flat YAML views: one file per INI section family under --out
  (`settings_<section>.yaml`), values verbatim plus `source:` coordinates;
  byte-identical on rerun (AC 2, TestSettingsViewsRegenerateByteIdentical).
- Flags: --ini, --out, --dry-run; parse failure of the ini fails the run
  before any write (exit 1); missing ini fails with the file named.

Open question resolution: one flat YAML per section family; no --single-file
variant (ponytail). Shortcut recorded: only engine-exposed typed seams are
checked, the full upstream option registry is not ported yet.

Gates at pushed head 2d2fd12 (worktree .worktrees/feat-migrate-settings):
go build OK; gofmt -l empty; go vet OK; ./scripts/lint.sh ./... => 0 issues;
go test ./... => 46 ok / 0 fail; just sqlc-check OK;
beans-lint OK (183 beans).

## Summary of Changes

Lane feat/migrate-settings (session task-dgmigratesettings-coder, 102 calls / 0 nudges; fix wave task-dgmigratesettings-fix-coder):
- cmd/devilutionGo-migrate/settings.go (498 lines) - the ADR-014 settings subcommand: reads diablo.ini through internal/engine's INI codec (reused, no parser reimplementation), runs validity checks, carries the file over BYTE-IDENTICAL into the engine settings dir, and writes one flat YAML registry view per INI section family (settings_<section>.yaml, values verbatim + source: coordinates).
- cmd/devilutionGo-migrate/settings_test.go (288 lines, 9 tests + resolution pins + BOM pin) - synthetic text fixtures, no binaries.
- main.go wiring + usage; saves lane coverage untouched; per-PR commit chain 2d2fd12, merge-ff92923, fix round cbfbed7/41f5668/bc51a09/41931c3.

## Proof of Work (parent verification, 2026-11-09)

- Pushed head 41931c30ef6a5b18055512aec5f1b3880e36b4e1 on origin/feat/migrate-settings; parent re-ran THE SAME gates at that head:
  - `go test ./cmd/devilutionGo-migrate/... -count=1` -> 130 pass
  - `go test ./... -count=1` -> 11,457 pass / 38 pkgs, 0 FAIL
  - `gofmt -l .` -> empty; `./scripts/lint.sh ./...` -> 0 issues; `beans-lint: OK (183 beans)`
- AC1 evidence: TestSettingsCarriesIniByteIdentical (cmp-level byte-identity via re-read bytes.Equal); TestSettingsReportListsViolationsItCannotFix (Game.Speed=99, Audio.Sample Rate=1234 + untyped notices listed, carry unaffected); parse failure exits 1 before any write.
- AC2 evidence: TestSettingsViewsRegenerateByteIdentical (byte-equal across two runs, one view per section family).
- AC3: all gates green at the pushed head; no save binaries in the diff (parent read the diff list).
- Review: brainer session task-dgmigratesettings-review-brainer, verdict REQUEST-CHANGES (.hermes/lane/dg-migrate-settings/dgmigratesettings.review.md) - 7 findings; fix wave landed all seven; parent operator checks re-run at the fix head (numbers above).

## Deviations recorded (fidelity gate)

- D1 (review F1/F2): upstream resolves the settings dir as "a WRITEABLE diablo.ini in the exe folder" (Source/utils/paths.cpp:116-118); this tool implements ADR-004's portable.ini marker law instead (comment now cites ADR-004, not the upstream line). Portable marker = recorded deviation per the ADR-004 record wording; upstream behavior restored if the ADR changes.
- D2 (review F2): the pref-dir fallback resolves the Linux SDL layout only (~/.local/share, XDG_DATA_HOME honored); macOS/Windows resolve differently upstream (Source/utils/paths.cpp:114/:135) - per-platform resolution is out of scope for this tool on this host; marked POSIX-only at resolveSettingsDir.
- D3 (ponytail, recorded): typed validity checks cover only the engine-exposed seams (Game.Speed bounds, Audio sample-rate set); the full upstream options registry is a future upgrade path (review F5-related, scoped).

## Residual

None for this bean. `_brainer triage owed` on nothing; the settings YAML view shape is an implementation decision (bean pins regeneration, not field names), recorded in the review file.
