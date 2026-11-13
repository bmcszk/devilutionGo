---
# devilGo-xal8
title: 'ADR-008 note: input/animation homes + gameplay shim deprecation'
status: completed
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T17:00:47Z
parent: devilGo-ok51
---

---
# devilGo-xal8
title: 'ADR-008 note: input/animation homes + gameplay shim deprecation'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:25Z
parent: devilGo-ok51
---

---
# devilGo-xal8
title: 'ADR-008 note: input/animation homes + gameplay shim deprecation'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:10Z
parent: devilGo-ok51
---

---
# devilGo-xal8
title: 'ADR-008 note: input/animation homes + gameplay shim deprecation'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# ADR-008 note: input/animation homes + gameplay shim deprecation

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-5).

## Goal
One ADR-008 paragraph recording the intended homes BEFORE the k736 wiring
wave picks paths (physical move deferred to the extraction milestone):
stick math -> internal/game/input post-wiring (thresholds as parameters;
no pcurs/PLR_MODE imports); AnimationInfo -> internal/engine/animation at
extraction (upstream home engine/animationinfo); pcurs/panels/PLR_MODE
semantics stay in diablo/controls; combat stays diablo;
internal/gameplay (re-export shim outside the three layers) deprecated —
callers import diablo/* directly.

## Acceptance criteria
- [x] ADR-008 paragraph names all five homes + the shim deprecation
- [x] No file moves in this bean

## Triage

Triage source: brainer session task-techdeptTriage-brainer (2026-11-12,
.hermes/lane/dg-chore/techdept.triage.md) — verdict CONFIRMED with the
scope noted above; parent filed this child from that verdict.

## Libraries

- stdlib only (no third-party deps at this stage)

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md — read fully before acting
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest correct change
/home/blaze/.agents/skills/go-dev/SKILL.md — repo conventions for docs edits
/home/blaze/.agents/skills/go-tests/SKILL.md — gate expectations

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Docs are rules, not narration: ADR edits record decisions with upstream
   citations; requirements.md must never reference paths that do not exist in
   the tree (the dgo-parity lesson). Simple English, active voice, no em
   dashes, labeled links.

## Proof of Work

2026-11-13: ADR-008 homes paragraph landed (commit 63de3d5), all five homes + gameplay shim deprecation, no file moves. Worker checked the Go tree before writing: internal/game/input and internal/engine/animation do not exist yet (no invented paths); AnimationInfo confirmed at internal/diablo/animation. Parent read the paragraph on the branch.
