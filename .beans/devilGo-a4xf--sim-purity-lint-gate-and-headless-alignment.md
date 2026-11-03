---
# devilGo-a4xf
title: Sim-purity lint gate and headless alignment
status: completed
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T16:51:03Z
parent: devilGo-xc37
---

# Sim-purity lint gate and headless alignment

## Goal

ADR-017 and the sim-purity discipline as a CI gate plus headless truth:
sim code (internal/game kernel, internal/diablo sim paths) must not import
rendering, ebiten, wall-clock time or file io; cmd/devilutionGo-headless
runs the kernel with a demo file end to end without a window.

## Research

- internal/game + internal/diablo import graph today (clean; keep it so).
- cmd/devilutionGo-headless exists with its own test run.
- scripts/lint.sh is the gate family home; coders run it pre-push.

## Scope

- scripts: new lint rule (import + time.Now/file-io ban list for sim paths).
- headless: kernel demo replay path proven in tests; no window, no GL.

## Acceptance criteria

- [x] scripts/lint.sh fails when a sim package imports rendering or ebiten.
- [x] headless binary replays a demo file with no display; go test ./... green.

## Libraries

- None new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md

## Dependencies & blockers

- Kernel core at minimum; best after tick log so headless replays are real.

## Open questions

- None.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.

Claimed 2026-11-03: lane dispatched, branch dg-purity-lint, base 8a404ae (PRs #24/#25 merged). Parallel-lane file-ownership law active.

## Proof of Work

### Commits (branch dg-purity-lint, pushed)
- c8fe434 devilGo-a4xf: simulation purity gate in scripts/simpurity wired into lint.sh
- ea925d1 devilGo-a4xf: headless demo replay proof with no display dependency

### Files
- scripts/simpurity/main.go (new): gate. Parses non-test .go sources of every dir
  under internal/game + internal/diablo (testdata pruned) and reports banned
  rendering imports (hajimehoshi/ebiten*, any path with a `render` segment),
  banned file io imports (os, io/ioutil) and wall-clock calls (time.Now, time.Tick,
  time.NewTicker, time.Since). Exit 0 clean, 1 violations, 2 scan failure
  (fail-closed). Output line: `<pkg>: <file>:<line>: <reason> imports "<path>"`.
- scripts/simpurity/main_test.go (new): negative proof over the fixture, _test.go
  exemption proof, clean + testsupport controls, ban-classifier table, fail-closed
  proof for a missing tree.
- scripts/testdata/impure/{impure.go,impure_test.go} (new): synthetic fixture with
  all four bans (ebiten, repo render path, os, time.Now); the `_test.go` half
  reads the disk and must stay unreported. `//go:build ignore` keeps the fixture
  out of the Go build/lint graph while the gate still parses it.
- scripts/testdata/clean/clean.go (new): purity control (time.Duration math only).
- scripts/testdata/exempt/testsupport/loader.go (new): fixtures-loader control.
- scripts/lint.sh (edit): runs `go run ./scripts/simpurity` repo-wide before the
  golangci-lint step, so the gate is part of every lint run.
- cmd/devilutionGo-headless/main_test.go (edit): TestReplay_DemoFile_HashesIdentically
  writes a demo file (engine/demo writer), replays it twice through
  game.NewDemoReplay(kernel.New(...)) at seed 1 and requires one trace hash and
  one event count; TestHeadlessBinary_HasNoDisplayDependency reads
  `go list -deps -f '{{.ImportPath}}' .` and fails if the headless graph gains
  ebiten or a render package.

### Gate output (final run, tree clean)
```
$ go build ./...
build rc=0
$ go test ./... -count=1
ok  github.com/bmcszk/devilutionGo/scripts/simpurity  0.002s
(all packages ok, no FAIL lines)
$ gofmt -l .
(empty, 0 files)
$ ./scripts/lint.sh ./...
simpurity: internal/game internal diablo clean
0 issues.
$ go test ./cmd/devilutionGo-headless/ -count=1 -run 'Replay|Display'
--- PASS: TestReplay_DemoFile_HashesIdentically (0.00s)
--- PASS: TestHeadlessBinary_HasNoDisplayDependency (0.03s)
```

### Negative proof (gate names package + file + offender, non-zero exit)
```
$ go run ./scripts/simpurity ./scripts/testdata/impure
scripts/testdata/impure: scripts/testdata/impure/impure.go:9: file io imports "os"
scripts/testdata/impure: scripts/testdata/impure/impure.go:12: rendering imports "github.com/bmcszk/devilutionGo/internal/engine/render"
scripts/testdata/impure: scripts/testdata/impure/impure.go:13: rendering imports "github.com/hajimehoshi/ebiten/v2"
scripts/testdata/impure: scripts/testdata/impure/impure.go:22: wall clock time.Now
simpurity: 4 violation(s) in ./scripts/testdata/impure
exit status 1
```
Clean tree run (defaults): `simpurity: internal/game internal/diablo clean` (rc 0).

### Scope note (internal/ untouched)
Three non-test internal packages use `os`: `internal/diablo/testsupport/*.go`
(fixture loaders for tests) and the orphaned `internal/diablo/gamedata/audit`
(manifest loader, no callers). Nothing under internal/ was edited (read-only for
this lane). The gate exempts them explicitly in `isFixtureLoader` (a path element
named `testsupport`, plus the one audit path) with a 2-line justification; if the
parent wants zero exemptions, `audit` should move under a testsupport-style
package or be deleted.

### Decisions recorded
- Wall-clock rule bans only the four named readers; `time` stays importable so
  `time.Duration` arithmetic (tick length) is legal in sim code.
- `os`/`io/ioutil` are banned at import level, which covers `os.*` file io and
  also keeps env/process access out of the sim.


### Follow-up commit ce2a24e (fail-open fix)
`simpurity` now counts the sim sources it parsed and prints the count in the pass
line; a tree with zero sources exits 2 instead of reporting clean, so a
misconfigured or empty root can never read as a pass. New test:
TestRun_EmptyTree_FailsClosed over `scripts/testdata/empty`.

```
$ ./scripts/lint.sh ./...
lint: sim purity gate
simpurity: internal/game internal/diablo clean (169 sources)
0 issues.
```
