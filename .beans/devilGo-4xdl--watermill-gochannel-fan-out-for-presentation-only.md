---
# devilGo-4xdl
title: Watermill GoChannel fan-out for presentation only
status: completed
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T16:48:08Z
parent: devilGo-xc37
---

# Watermill GoChannel fan-out for presentation only

## Goal

ADR-018 laws 2 and 5 as code: after each ordered pass the kernel publishes
its events on the in-process bus for presentation consumers (renderer, sound,
UI flourishes). Sim-mutating handlers never subscribe to the bus; the kernel
calls them itself. This adds watermill to go.mod exactly as the catalog row
already records it: in-process dispatch, backend swappable, presentation only.

## Research

- docs/go-libraries.md catalog entry for the bus (GoChannel backend note).
- Watermill reference research from the architecture wave: default publish is
  non-blocking; cross-subscriber order is not guaranteed; therefore the kernel
  owns dispatch and the bus only fans out display traffic.
- Presentation consumers today: cmd/devilutionGo UI assembly, future sound
  catalog work.

## Scope

- go.mod addition per catalog; kernel publishes events post-pass; thin
  subscriber helper for presentation packages.
- Tests: zero-subscriber kernel unaffected; with N subscribers the ordered
  pass result is bit-identical; a content-change attempt from a subscriber
  is rejected by test guard.

## Acceptance criteria

- [x] Given 0 or 3 subscribers, the kernel pass produces identical state and event trace.
- [x] go.mod carries watermill matching the catalog row; go.sum clean.

## Libraries

- ThreeDotsLabs/watermill (catalog; the only new dependency in this feature)

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md

## Dependencies & blockers

- Kernel core + tick log children first (bus consumer has nothing to read
  until events exist).

## Open questions

- None; sync fan-out inside the pass stays forbidden by law 3.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.

Claimed 2026-11-03: lane dispatched, branch dg-watermill-fanout, base 8a404ae (PRs #24/#25 merged). Parallel-lane file-ownership law active.

## Proof of Work

### Chosen dependency
- `github.com/ThreeDotsLabs/watermill v1.5.3` (newest stable v1 under v1.5.3 per
  `go list -m -versions github.com/ThreeDotsLabs/watermill`; v2 not offered).
- Direct require in `go.mod`; `go mod tidy` run, `go mod verify` → all modules verified.
- Only new direct module. Transitive (indirect, from watermill's own go.mod):
  `lithammer/shortuuid/v3 v3.0.7`, `oklog/ulid v1.3.1`, `pkg/errors v0.9.1`;
  `golang.org/x/sys` bumped 0.22.0 → 0.44.0 because watermill requires ≥0.44.0.
  `go.sum` additionally records module-graph-only lines (testify/go-spew/uuid 1.2.0).

### Commits
- a11e5c0 `devilGo-4xdl: publish kernel pass events on an in-process watermill bus`
  (pushed to origin/dg-watermill-fanout).

### Files
- `internal/game/kernel/kernel.go` — `Publisher` interface + `SetPublisher`;
  `Pass` fans out `slices.Clone(events)` only after the trace append, only when
  the pass emitted something. No publisher ⇒ byte-identical behaviour to pre-bus.
- `internal/game/bus/bus.go` (new) — GoChannel-backed `Publisher` (one message
  per pass, JSON payload), `Subscribe(ctx)` helper returning decoded
  `[]kernel.Event` batches, `Close`.
- `internal/game/bus/bus_test.go` (new) — 3 tests.
- `go.mod`, `go.sum`.

### Verified backend semantics (watermill v1.5.3, pubsub/gochannel source read)
- `Publish` returns before delivery unless `BlockPublishUntilSubscriberAck` is
  set (we leave it off) — the simulation pass never blocks on a consumer.
- Each subscriber is fed from its own goroutine and each receives a message copy
  (`msg.Copy()`), so cross-subscriber and cross-pass delivery order is NOT
  guaranteed; the kernel trace stays the authoritative order. Encoded as a 2-line
  caveat on `bus.Publisher` + `Subscribe`.
- A subscriber must `Ack` (else that subscriber's send loop blocks) and messages
  published with no subscriber are dropped (non-persistent) — we ack always and
  never retry undecodable display traffic.

### Acceptance criteria
- [x] 0 or 3 subscribers ⇒ identical state + event trace: table test
  `TestPass_SubscribersDoNotChangeThePassResult` compares `noBus` /
  `busZeroSubscribers` / `busThreeSubscribers` against the pre-bus reference by
  `TraceHash`, full `Trace()` and per-pass returns; each of the 3 subscribers
  must also have received exactly the reference trace.
- [x] watermill in `go.mod` matching the catalog row; `go.sum` tidy.
- [x] Publish-after-pass ordering: `TestPass_EventsArePublishedOnlyAfterThePassCompletes`
  fails if a handler observes a delivered event while dispatching, and requires
  each pass's event only after `Pass` returns.
- [x] Mutation guard: `TestSubscribe_MutatingReceivedEventsCannotLeak` mutates the
  whole received batch (Kind/A/Tick) and fails if the kernel hash/trace or the
  second subscriber changed.

### Gates (all green, run at HEAD a11e5c0)
```
## gate: go build ./...
rc=0
## gate: go test ./... -count=1
40 packages ok, 0 FAIL
## gate: gofmt -l .
(empty)
## gate: ./scripts/lint.sh ./...
0 issues.
```
Also `go test ./internal/game/... -race -count=1` → all ok.

### Notes / deviations
- `Game`/`cmd` wiring is out of lane ownership: the seam is `kernel.SetPublisher`,
  which the UI assembly (cmd, another owner) calls once per process.
- No presentation subscriber is registered in-tree yet — the helper is the
  deliverable; first consumer is the renderer/sound wiring bean.
