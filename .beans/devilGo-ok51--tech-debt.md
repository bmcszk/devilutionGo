---
# devilGo-ok51
title: Tech debt
status: in-progress
type: epic
priority: high
created_at: 2026-11-12T23:56:03Z
updated_at: 2026-11-13T17:00:46Z
---

---
# devilGo-ok51
title: Tech debt
status: in-progress
type: epic
priority: high
created_at: 2026-11-12T23:56:03Z
updated_at: 2026-11-12T23:56:15Z
---

# Tech debt audit (2026-11-12, parent session, honest)

Parent code audit at master 838af54 (+ PR #66 branch) against the 21 ADRs,
the beans ledger, and the user's four questions. Everything below is
verified against the tree (grep/read), not remembered.

## Verified-good (so the debt list is honest, not doom)

- Layering (ADR-008) holds: zero `game -> diablo` imports; one exception
  found (internal/light/tables.go imports diablo/gamedata/leveldat) —
  filed below as TD-1.
- ADR-020 two binaries: mechanically gated, green. ADR-021 one tool
  binary: now holds after the dgo-parity fold.
- ADR-018 event core EXISTS and is real: kernel (commands in/events out,
  Gate, ordered pass, trace + TraceHash), tick log in SQLite
  (tick_commands/tick_events/tick_checksums, snapshot cadence constant),
  watermill bus for presentation fan-out. Headless mode runs the same
  kernel (ADR-017).
- Persist: sqlc law respected; .sv importer present (ADR-003/011).
- Gates: TDD RED-first culture, lint executable-surface + sim-purity +
  whenthen stages, parity harness slice 1. 47 packages green.

## Tech debt items

### TD-1 [layering] internal/light imports internal/diablo/gamedata/leveldat
internal/light/tables.go:7 pulls Diablo level data into a layer the ADR-008
map treats as below-diablo (light serves both). Fix: invert the dependency
(leveldat-independent light table API + a diablo-side adapter), or move
light under diablo and keep a generic lighttable math package generic.
Small, but it is THE pattern that rot starts from.

### TD-2 [event coverage] The merged input lanes are event-dead-ends
diablo/controls (stick core + plrctrls action machine) emits DECISIONS
through func hooks; nothing submits them as kernel Commands. Kernel carries
only 18 kinds (input trio + PlayerMove/PlayerAttack + debug); the action
machine's whole dispatch family (SpellSelect/TryIconCurs/CheckTriggers/
store release paths) has NO command kinds, NO gate rules, NO events.
Per ADR-018 law 1 (everything changes state through commands) the wiring
lane (k736) must bridge controls decisions -> kernel commands, and the
action families need kinds + gates. Until then the event-sourcing claim is
true for the kernel but vacuous for gameplay.

### TD-3 [event coverage] Gameplay packages bypass the kernel entirely
internal/diablo/inventory, stores, combat, missiles, entities mutate state
by direct calls; zero kernel.Event or bus.Publish usage. They are ports
with honest fidelity, but they are NOT event-sourced: no ItemPickedUp,
DoorOpened, SpellCast kinds exist despite ADR-018 naming them. The replay
story ("the log is the game", law 4) currently covers movement + debug
only. Debt: define the per-domain command/event kind sets as each domain
wires, or accept an explicit carve-out note in the ADR for pure-port
packages until their wiring wave.

### TD-4 [event-sourcing] Tick log vs .sv byte-compat dual truth
diablo/save (.sv codec, byte-compat) and persist/ticklog are two save
stories. ADR-003/011 say SQLite-only live saves with importer; ADR-018
says log-as-save. Nobody has written down which wins for a
mid-dungeon session restore (replay from snapshot vs .sv import).
Needs an ADR-level decision note (one paragraph in ADR-003 or -018).

### TD-5 [split] controls package sits under diablo but is generic
diablo/controls is an input DECISION core keyed on upstream C++ names;
nothing in it reads Diablo tables. The generic/Diablo split (user question
4) would place the decision math under internal/game/input (or engine) with
diablo/controls as the thin upstream-fidelity adapter. Same question applies
to diablo/animation and diablo/combat skeletons: port-fidelity lives in
names, but the MECHANISMS are generic ARPG. Not urgent; record the intended
home before more input/ui lands there.

### TD-6 [gaps] Known-unimplemented surfaces (tracked, listed for honesty)
- zdj7 automap renderer + qeyu automap primitives: unstarted (7aul's
  automap-on golden depends on them).
- k736 input wiring (ebiten axis/keys -> controls -> kernel): unstarted;
  the stick/action machine is unreachable code until it lands.
- Networking (ADR-016 loopback): internal/net does not exist yet.
- Lua hooks (ADR-006): no stub package yet; decision says stubbed-now,
  nothing exists.
- SaveGameData full-stream assembly (loadsave.cpp:2769-2935 walk):
  the c6wt session parity slice named 16 upstream sections; Go covers
  hero/header/grids/dropped-items only. Level-stream writers (monsters/
  quests/missiles/lights/vision/premium) missing.
- Parity harness: slice 1 only (comparator+capture). C++ dumper, real
  .dun golden frames, CI wiring all open (7aul scope b-d).

### TD-7 [drift risk] Upstream anchor moved
Upstream pins cite ca91d778f; devilutionX checkout HEAD is e1ad0c2b6.
The ADR-002 fidelity mandate needs a re-pin decision: either re-pin the
contract anchor set to e1ad0c2b6 (diff the touched hunks) or record why
ca91d778f stays canonical. Undefined anchor drift will bite parity goldens
later.

### TD-8 [tests] tick-log restore path unproven end-to-end
TickLogWriter + checksums exist with unit tests, but no test does
save -> close -> reopen -> replay-from-snapshot -> continue-and-hash-match
on a real store file. That is the ADR-018 law 4 load-bearing promise.
One integration test closes it.

### TD-9 [docs] docs/requirements.md + research still cite dgo-parity
The fold moved the CLI; two docs mention the old path (research only,
non-binding). Sweep in the next docs-touching lane.

## Deliberately NOT debt
- Entity/exported-symbol doc coverage, comment style: audited by lint gates,
  green.
- The partial persist roundtrip (hero/header/grids/items): honestly named in
  the c6wt ledger, tracked above as TD-6, not silent debt.
- Bean ledger hygiene: 185 beans lint green; statuses match merged truth
  after the post-#65 reconciliation.

## Triage outcome (brainer task-techdeptTriage-brainer, 2026-11-12)

Full verdict: .hermes/lane/dg-chore/techdept.triage.md. Children filed:

- [ADR-002 note: parity anchor ca91d778f canonical](.beans/devilGo-mxda--adr-002-note-parity-anchor-ca91d778f-canonical-no.md) (high, micro) — TD-7
- [ADR-003 note: save-story precedence](.beans/devilGo-bixn--adr-003-note-save-story-precedence-rows-save-tick.md) (high, micro) — TD-4
- [ADR-008 note: input/animation homes + gameplay shim deprecation](.beans/devilGo-xal8--adr-008-note-inputanimation-homes-gameplay-shim-de.md) (normal, micro, BEFORE k736 wiring picks paths) — TD-5
- [Tick-log session-resume integration test](.beans/devilGo-gvmc--tick-log-session-resume-integration-test-closereop.md) (high, micro) — TD-8
- [Invert light to leveldat](.beans/devilGo-367i--invert-light-to-leveldat-light-owns-its-leveltype.md) (normal, micro) — TD-1
- [Docs sweep: retire dgo-parity citations](.beans/devilGo-ve6v--docs-sweep-retire-dgo-parity-citations.md) (low, micro) — TD-9
- TD-2 -> [k736 child: controls decisions to kernel commands](.beans/devilGo-iq0d--k736-child-controls-decisions-to-kernel-commands-e.md) (critical, full lane, filed under the input epic) — the M1-critical event bridge
- TD-3 -> owned by the [gameplay epic](.beans/devilGo-nxxo--gameplay-combat-spells-missiles-items-inventory-stores.md) (its ADR-alignment note already commits the per-slice conversion); epic links, no child
- TD-6 -> all rows owner-verified (zdj7, qeyu, k736, zo8j, yjuv/yd48, 7aul, c6wt open AC); epic links only

Corrections the triage made to this body: kernel carries 14 kinds (not 18);
the SaveGameData bullet was stale — monster/quest/missile writers exist
(oqho completed); genuinely open: SaveLighting/vision block, premium-item
section, game-layer session snapshot (owned by the cursor bean's open
session AC). Upstream checkout HEAD sits on a fork FEATURE branch
(e1ad0c2b6) — ca91d778f stays the canonical anchor (C1 records why).

Dispatch order: C1+C2+C3+C4+C5 (+C7 opportunistic) as micro chores; C6 as
the next full input lane once its c6wt dependency check passes.

## Progress log

Claimed 2026-11-12 (parent audit session): the audit body above IS the
claim evidence; each TD item carries its file:line verification from the
master 838af54 tree. Items leave this epic by becoming bean children or
landing as fixes; the epic closes when the list is empty.

## Progress log

Claimed 2026-11-12 (parent audit session): the audit body above IS the
claim evidence; each TD item carries its file:line verification from the
master 838af54 tree. Items leave this epic by becoming bean children or
landing as fixes; the epic closes when the list is empty.

## Progress log — 2026-11-13 one-by-one execution

- TD-7 -> mxda CLOSED (14f3794): ADR-002 anchor section + drift table; worker corrected the Interact drift number (:532 not :533) by grepping.
- TD-4 -> bixn CLOSED (53f15e8): ADR-003 precedence + ADR-018 law-4 cross-ref; upstream anchors verified, demomode.cpp full path corrected.
- TD-5 -> xal8 CLOSED (63de3d5): ADR-008 five homes + shim deprecation; worker verified game/input and engine/animation don't exist yet (no invented paths).
- TD-9 -> ve6v CLOSED (786ccb7): 6 live dgo-parity citations fixed; superseded notes; historical records untouched.
- TD-8 -> gvmc CLOSED (775a49d) with HONEST premise falsification: the close/reopen/replay/continue spine already worked; TestTickLogSessionResume (test-only, +209/−26) now pins ADR-018 law 4 instead of assuming it. Contract RED clause exercised correctly (STOP, no weakening).
- TD-1 -> 367i CLOSED (5c6f84d): light owns LevelType; leveldat import GONE; parse-based layer-pin test; adapter in internal/diablo/levels/light.go; goldens byte-unchanged (parent re-ran them).
- Remaining: TD-2 -> iq0d (controls->kernel bridge, full lane, M1-critical, gated by c6wt internals); TD-3 -> owned by nxxo per-slice conversion; TD-6 -> owner beans.
Branch chore/techdebt-fixes @ 5c6f84d (5 commits on b000c0d) ready for parent PR on user go.

## Progress log — 2026-11-13 TD-2 closed

- TD-2 -> iq0d CLOSED (feat/eventbridge, 7 commits c54e6b7..ba768ec,
  parent-verified): the event bridge is REAL — decisions submit as
  gate-checked commands, replay is checksum-verified. Two review-caught
  defects fixed (layering move to diablo/inputbridge; overlay gate made
  real). Remaining epic items: TD-3 (nxxo-owned, per-slice), TD-6
  (owner beans) — epic continues as those lanes land.
