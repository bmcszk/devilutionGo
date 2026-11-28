---
# devilGo-ok51
title: Tech debt
status: in-progress
type: epic
priority: high
created_at: 2026-11-12T23:56:03Z
updated_at: 2026-11-28T15:36:15Z
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

## Progress log — wrap-oracle slice (2026-11-13, session 2)

- PR #70 merged as 9a1597e161: wrap-oracle goldens landed (od1d lane).
- Lint blind spot found (absolute-path dead skill refs slip the path check): 4 beans repointed (bh4m/dufb/qr5a/y1y9), filed as [beans-lint: skills-path check misses absolute-path dead refs](.beans/devilGo-n63m--beans-lint-skills-path-check-misses-absolute-path.md) (draft, p1t8 child).
- Stale chore/postmerge58-custody branch deleted after parent adjudication (#59 squash carried the scrub).
- Remaining epic items: TD-3 (nxxo per-slice), TD-6 (owner beans); v1hy keeps integration-hash + atlas-per-size open; next functional lane: c6wt SaveGameData stream assembly (grouped-PR wave).

## Progress log — hl1 hero-load lane triage (2026-11-25)

Brainer task-hl1-triage over dcob/xoap/py4t/h6co, parent-verified (4/4 spot-check cite sets confirmed on 949c728). Findings: RF1 (created heroes never settle in the shipped binary) confirmed open and unowned — now dcob D10; slot-0 overwrite defect filed (dcob D3); upstream Load Game / New Game path absent in Go — split to a new draft bean blocked-by dcob; ADR-003 store-model contradiction resolved by parent decision (ADR-003 per-hero files stand; boot single-db migration rides the D9 bean). xoap/py4t/h6co promoted to todo with lint-complete skill sets; wave feat/m2-hero-load dispatched (9-commit order, D9 out of wave). Evidence: .hermes/lane/m2-hero-load/result.md.

## Progress log — hl2 v13s triage (2026-11-25)

Brainer task-hl2-load-triage over the load bean (parent-verified: persist/doc.go import direction, loadSlotHandler seed gate, no FrontGame transition, zero-pack rows). v13s re-scoped to 8 slices (S1-S8), adapter home decided internal/persist (deviation recorded), D1 enumeration migration split to a new draft bean devilGo-lz2m blocked-by v13s. Six deviations recorded. Wave feat/m2-hero-load-v13s dispatched (5 commits).

## Progress log — hl3 lz2m triage (2026-11-25)

Brainer task-hl3-mig-triage (parent-verified upstream layout). lz2m re-scoped to 4 slices (S1 HeroStoreSet in persist, S2 SessionArchive slot routing, S3 boot+screen rewiring, S4 first-free/cap/legacy re-home sweep), promoted todo, wave feat/m2-hero-load-lz2m-per-hero-store dispatched. Upstream per-slot file layout verified live (no handle map; open-per-operation).

## Progress log — hl4 gm triage delta (2026-11-25)

Brainer task-hl4-gm-triage (parent-verified). Game-menu bean re-scoped to AC1-AC13 (6 slices S1-S5 + Options follow-up bean filed). Corrections: in-game Options = in-engine slider table (not DiabloUI SettingsMenu); tick rate 20-50 is nTickRate (gnTickDelay=1000/rate); SimOptions already exists unconsumed; KindQuit reused; PauseMode deferred; Save/Load entries real. Wave feat/1a4r-game-menu dispatched (5 commits).



## Progress log — m2-followups wave (2026-11-26)

Brainer triage over uscq (persist sweep honesty + ghost pin) and 7bij (screenshot ReadPixels panic) — both PROMOTE-TODO, parent spot-checked every cite (line numbers, upstream pfile.cpp:693-706, the ebiten/rendering SKILL.md:41,49 misguidance). Wave fix/m2-followups dispatched (task-m2fu-wave-coder): 2 commits 07a7615 + d4356f1, TDD honest (uscq falsification recorded, 7bij real RED panic captured), parent gates + clean-clone oracle green at d4356f1, independent review APPROVE (no blocking; NB-1 -> new draft bean, NB-2 wording note in the uscq PoW). Skill-text defect filed as its own draft under p1t8. Wobble: the wave branch was cut before the promotion ledger commit landed (branch carried stale draft-bodied bean copies) — repaired by a402813 chore commit to master + merge 1a066fb into the branch; beans CLI then closed both beans on the merged copies. Environment repair: dead pipx orphan shim ~/.local/bin/python3 (headroom-ai venv uninstalled) was shadowing python3 for every shell incl. the pre-push hook's beans-lint — shim removed, gate verified green through the hook's own resolution path.



## Progress log — gate-parity wave (2026-11-26)

User order: git hooks - lint on commit, test on push, same checks in gha, all coding-rules rules and more. Parent audit (.hermes/lane/gate-parity/audit.md) found three pre-push variants, a paper-gate installer, missing vet/windowed/no-raw-sql/beans-lint/sqlc in hook+CI, and go-version drift. Brainer triage (task-gp-triage) corrected two audit claims (W1 third variant, W2 AC4 smoke not skipping asset-less -> in-bean enabling fix S0) and resolved D1-D9. Wave chore/gate-parity (task-gp-wave-coder): 5 commits + repair round after review round 1 REQUEST-CHANGES (B1 coder-side AC ruling -> parent DECIDED option b; B2 tautological cmp -> committed-source cmp + mixed-checkout proof); review round 2 APPROVE. 13/13 ACs ticked, parent gates green at cd20c42, ledger custody on the branch. Gate proof protocol: seeded failures in scratch clones, canary PR #88 closed, no hook bypassed.



## Progress log — frame-image helper wave (2026-11-26)

NB-1 follow-up from the gate-parity review landed: cha7 promoted from the reviewer's own scoping (three file:line cites; home decision exported-from-scene over a new testsupport package), one-commit refactor wave refactor/frame-image-helper (task-frame-img-coder, 59bdb83, net -32), mechanical behavior parity (identical test-count events both heads), brainer APPROVE with stdlib-png-writer-grade semantics audit, parent gates + oracle green. Public-API export (scene.FrameToImage) reviewed per the risky-leaf law.



## Progress log — framebudget hardening wave (2026-11-26)

n5an promoted as the fix unit (triage decision: min-of-2 child passes, budget + factor untouched; recalibration rejected with knife-edge math; path corrected internal/engine -> internal/integration). Wave test/framebudget-harden (task-framebudget-coder, 95d6c24): budgetClass.passes + measureRowPasses fail-only-if-all-miss with per-pass quoting; scripted pass-rule matrix (the discriminating fail-then-pass walk RED on pre-fix, mutation flip proven); local path byte-identical. Review APPROVE (gate-semantics strict round). PR #91 CI green (run 45881359741 = AC3's first-run evidence; pass-log visibility deviation recorded in the bean PoW).



## Progress log — status-oracle + deletehero pin waves (2026-11-26)

Two parallel small waves off master f6ac11e: 2yge (status.sh raw golangci-lint -> the pinned scripts/lint.sh oracle; seeded FAIL/clean proof; APPROVE) and ym0r (DeleteHero atomicity pin: the weak PR14-era atomic test upgraded to a four-case deterministic injection matrix; BeginTx-removal mutation reds 3 of 4 cases, one honestly vacuous; APPROVE). Both parent-gate-verified at their heads. Bookkeeping notes: beans create with -d - blocks without stdin (use --body-file); 2yge initially shipped without the beans-lint-mandated ponytail skill line - the pre-push hook caught it on the parent's own push (the new gate biting its author, twice now).



## Progress log — Options-slice wave (2026-11-26)

First epic-sized PR under the new sizing law (user: PRs per epic or ~8h human work; skill law recorded in devilutiongo-lane-ops): THREE beans one wave (ow6p audio + t4fz keymapper data half + w0ky options rows), ONE triage (task-opts-triage-brainer, 660-line evidence file), FOUR commits + custody, ONE strict review (9 claims, APPROVE static, two triage corrections: QuickMessages=10 not 16, TSV 58 rows measured). E-A1 answered by measurement: retail WAVs are linear PCM 22050 16-bit (probe, asset-gated, STOP rule unused). Gamma independent AC landed; presenter recomposes on brightness; boot brightness now ini-driven (disclosed darker default). Deferrals filed: devilGo-fxui (padmapper), devilGo-hi7t (keymapper callbacks). Parent battery: 14 checks green, one mutation re-proven live.



## Progress log — parity-close wave (2026-11-27)

THE GOAL epic (1n0a) driven to near-closure: combined triage ruled PROMOTE-PARTIAL (i2cr landable, yuqk already-adopted, uc44 split - the timedemo .sv oracle honestly blocked on the kernel epic: NO production path can produce a level-up, XP is pure-function + test handlers only). Epic body's 2026-11-03 findings corrected by evidence (384 goldens exist; CI budgets exist - the benchmarks AC ticked after a contention-failed first run re-ran green twice). Wave test/parity-close: `just parity` named+timed (18-22s, 14x headroom, D1 single-source via justfile into ci.yml+pre-push) + the closed-world data audit (96 files, 0 problems, fingerprints vs parsed Go rows, RED-first). 15 stale blocker edges cleared across i2cr/uc44. Remaining for 1n0a: the timedemo oracle (xc37 kernel) + state-hash-vs-C++ (same dependency). Review APPROVE (static, 12 claims); parent ran the seeded-failure + mutation proofs live.



## Progress log — goal2 wave (2026-11-27)

Goal lane 2: uc44's four infrastructure boxes + v791's data half, ONE wave (triage PROMOTE-PARTIAL). Real finds: (a) the Go store codec-decoded only 'hero' where upstream ReadArchive decodes every entry - fixed READ-only, the real reference save now loads (574799 bytes committed as text, hero entry 1288 -> 1266 = PlayerPackSize); (b) the tape-parser off-by-one was REAL but invisible to the old .hex test (different framing) - fixed with a hand-decoded behavioural RED; (c) the 32-byte upstream demo header needed the settings-block skip. v791: PlaceGroup pure port keeps the vanilla yp += deltaX bug verbatim (monster.cpp:341) with a scripted-RNG-tape oracle; parent re-proved THE parity mutation live (fixing the bug reds the pin: (15,18) vs want (15,16)). Sim wiring drafted as devilGo-h2zj. uc44 box 5 (byte-equal replay) remains the Goal's last gate - BLOCKED on the kernel epic (xc37): the sim must produce a real level-up state for WarriorLevel1to2. Fixture provenance pinned: SV 574799/78c0708a, DMO 79012/1c414e1b @ e1ad0c2b6, upstream tree clean.
