---
# devilGo-n63m
title: 'beans-lint: skills-path check misses absolute-path dead refs'
status: in-progress
type: bug
priority: low
created_at: 2026-11-13T18:45:37Z
updated_at: 2026-11-14T11:16:58Z
parent: devilGo-p1t8
---

The skills-path lint check caught ocl1's dead go-dev-notes ref (tilde form, hermes skills root) but misses the same dead path written absolute (/home/blaze/.hermes/skills/...). Four beans cited the dead path absolute (bh4m, dufb, qr5a, y1y9 - repointed to ~/.agents/skills/go-dev/SKILL.md by parent 2026-11-13); lint stayed OK both before and after, proving the blind spot. Fix: normalize ~ and absolute /home/blaze forms before the exists check, or match both. Evidence: postmerge676869 drill log + this fix diff.

## Triage (parent mini-triage, 2026-11-13)

1 root cause: scripts/beans-lint.py:276 path regex covers only (~/|/home/blaze/) for .agents and only ~/ for .hermes — an absolute /home/blaze/.hermes/skills/... token never matches, so dead refs in that form skip the exists check (4 beans slipped through; caught by hand during the postmerge70 sweep).
2 local: scripts/beans-lint.py:273-279 (the check), evidence in this bean's description; .beans/* affected files repointed 2026-11-13.
3 online: n/a (local script).
4 scope: extend the regex to match (/home/blaze|~) prefix for BOTH roots (one line), plus evidence: before/after lint runs against a scratch bean carrying the absolute dead path (RED -> GREEN, recorded in PoW). No other lint behavior changes.
5 acceptance criteria:
- [ ] Regex matches absolute-home dead paths (scratch-bean RED -> fix -> GREEN recorded in PoW)
- [ ] Full beans-lint OK on the real ledger after the change (no new findings beyond expected)
- [ ] Existing lint behaviors unchanged (tilde .agents + .hermes forms still caught; ocl1-class case still caught)
6 skills for coder: none needed (python lint script edit; no Go).
7 blockers: none. Joins the grouped SaveGameData wave PR (user grouping law).

## Libraries

- stdlib only (scripts/beans-lint.py is python; re/pattern only)

## Acceptance criteria

- [ ] Regex matches absolute-home dead paths (scratch-bean RED -> fix -> GREEN recorded in PoW)
- [ ] Full beans-lint OK on the real ledger after the change (no new findings beyond the scratch bean)
- [ ] Existing lint behaviors unchanged (tilde .agents + .hermes forms still caught)

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-dev/SKILL.md` — gates discipline
- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves the task

## Triage

Parent mini-triage 2026-11-14 (session _chat): root cause verified at scripts/beans-lint.py:276 (path regex covers only two of four root-spelling combos); scope = one regex line + RED-first scratch-bean evidence; joins the runnable-smoke grouped PR. Brainer re-review rides the PR review round per lane law.

## Progress log — smoke slice (2026-11-14)

- RED re-verified independently on this lane: reverted the :276 regex to
  the pre-#71 form (tilde .agents, /home/blaze .agents, tilde .hermes
  only), added scratch bean devilGo-scrat carrying
  /home/blaze/.hermes/skills/nonexistent-dir/SKILL.md -> lint exited 0
  ("OK (194 beans)", silent miss). Restored the current regex -> same
  scratch bean errors "skills path missing on disk" (exit 1). Scratch
  deleted; full lint OK (193 beans).
- The one-line fix itself had already landed upstream in #71 (this lane
  branched from 6ffeb2c which carries it); this lane supplies the RED
  evidence the AC asks for and confirms no further change needed.
- ACs satisfied: absolute-home dead paths caught (RED -> GREEN recorded
  above); full beans-lint OK on the real ledger; tilde .agents/.hermes
  and ocl1-class cases unchanged (same regex line, upstream-tested).
