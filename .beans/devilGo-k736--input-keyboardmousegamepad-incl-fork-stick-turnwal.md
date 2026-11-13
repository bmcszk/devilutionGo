---
# devilGo-k736
title: 'Input: keyboard/mouse/gamepad incl. fork stick turn/walk'
status: todo
type: feature
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-i56j
---

---
# devilGo-k736
title: 'Input: keyboard/mouse/gamepad incl. fork stick turn/walk'
status: todo
type: feature
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-11-03T12:52:00Z
parent: devilGo-i56j
---

# Input layer

## Goal
Kb/mouse first-class, gamepad with virtual cursor, keymapper/padmapper, and the fork's headline stick behavior.

## Research
controls/controller_motion.cpp:35 ScaleJoystickAxes (radial deadzone), :129-146, :192-205; plrctrls.cpp fork HEAD ca91d778f: Movement() Firm→WalkInDir / Weak→TurnToDir (:1861-1882), TurnToDir (:1776-1782), IsStickMovementSignificant (:1916-1920), Interact fake-attack via _pdir (:557-563); controller_motion.h GetLeftStickDirection unification; keymapper.hpp:17-18; per-frame hooks plrctrls_every_frame:2305 / after_game_logic:2312 (diablo.cpp:1568); ControlTypes scrollrt.cpp:228.

## Scope
internal/input: devices, deadzone scaler, StickPush classifier, plrctrls action machine, keymapper/padmapper with JSON config, auto control-mode switch.

## Acceptance criteria
- [ ] ScaleJoystickAxes port passes curve tests (deadzone, magnitude remap)
- [ ] StickPush thresholds: Weak turns in place (StartStand with new _pdir), Firm walks — behavior tests with synthetic axes
- [ ] Interact fake-attack uses current facing when idle (fork semantics)
- [ ] Keymapper defaults match C++ ini defaults; rebinding persists
- [ ] Mouse+kb and gamepad auto-switch active device (ControlMode)

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
/home/blaze/.agents/skills/ebiten/input/SKILL.md — axis/button/touch APIs
/home/blaze/.agents/skills/go-dev/SKILL.md — config persistence

## Dependencies & blockers
blocked_by dg-epic-01 only. Integrates with gameplay in epic-05 hooks.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DEFERRED — question answered by recorded milestone decision; revisit point documented in the question text above.
- Touch virtual gamepad deferred to mobile milestone (H12).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Input layer with the fork's headline stick behavior. Wired blockedBy qna9 only ✓ (body "dg-epic-01") — correctly unblocked early. 3 children.
2 local: Cites verified at fork HEAD ca91d778f (git log -1 matches bean's claim exactly): controller_motion.cpp:35 ScaleJoystickAxes ✓; plrctrls.cpp TurnToDir:1776 ✓, WalkInDir:1747 ✓, IsStickMovementSignificant:1917 ✓. Skills ✓.
3 online: n/a.
4 solution: none — ready. ACs testable with synthetic axes (curve tests, threshold behavior).
5 understanding: engine/input layer; fork semantics (Firm→WalkInDir / Weak→TurnToDir) are the differentiator vs upstream — correctly captured.
6 repro: `git -C ~/github/bmcszk/devilutionX log -1 --format='%h %s'`; `grep -n "TurnToDir\|IsStickMovementSignificant" Source/controls/plrctrls.cpp | head -4`
7 fix: none.
8 validation: children (t4fz, sf6k, c6wt) triaged in batch-08.
9 related: blocker qna9; integrates with nxxo hooks.
VERDICT: TRIAGED-OK

## TRIAGE 2026-10-18 (pi wave, batch 3)

devilGo-k736** (Input epic, parent i56j)
1 root: kb/mouse + fork stick semantics; gate qna9 ✓ wired (dg-epic-01); early-unblock intent correct.
2 local: HEAD `ca91d778f` matches ✓; `controller_motion.cpp:35 ScaleJoystickAxes` ✓; `plrctrls.cpp` WalkInDir:1747, TurnToDir:1776, Firm/Weak in Movement ~1861-1876 ✓, IsStickMovementSignificant:1917 ✓, fake-attack `_pdir` :557-563 ✓; `diablo.cpp:1568 plrctrls_after_game_logic` ✓. 3 children exist (t4fz,sf6k,c6wt).
3 online: SKIPPED.
4 solution: **epic Scope stale**: "keymapper/padmapper **with JSON config**" contradicts SETTINGS FORMAT AMENDMENT (diablo.ini 1:1); child t4fz already patched to INI (fix2 note) — epic body lags child. One-line scope edit needed.
5 DDD fit: internal/input→engine/input ✓.
6 dep-order: qna9 → k736; gameplay hooks integrate at nxxo. ✓.
7 AC-gap: ACs themselves fine (synthetic-axes testable); AC5 "ini defaults" already INI-correct — only Scope line conflicts.
8 validation: deadzone curve tests; StickPush Weak=turn/Firm=walk behavior tests; rebind persistence via diablo.ini.
9 related: qna9, nxxo, child t4fz (canonical INI fix).

**NEEDS-WORK:epic scope still says "keymapper/padmapper with JSON config" — contradicts diablo.ini-1:1 amendment; child t4fz already fixed, epic body not patched**


Wave outputs: /tmp/dg-triage/out-b3.md

## Libraries

Binds ebiten v2 for rendering (owner flag in Scope); everything else stdlib.


2026-11-03 ADR alignment: input becomes a command producer, no direct state
mutation. The kernel exposes the dispatcher under internal/game and both the
click router and the debug router feed its gate before the ordered pass; that
child is [Gate plus first gameplay handlers as commands](devilGo-noii--gate-plus-first-gameplay-handlers-as-commands.md); ([ADR-018](../docs/adrs/adr-018.md) laws 1-3).
Parent kernel: [Deterministic command/event simulation kernel (ADR-015, ADR-018)](devilGo-xc37--deterministic-commandevent-simulation-kernel-adr-0.md);

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
