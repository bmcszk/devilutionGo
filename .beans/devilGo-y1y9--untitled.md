---
# devilGo-y1y9
title: Fix garbled/stacked doc comments (lint-sweep fallout)
status: todo
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-10-24T17:26:11Z
parent: devilGo-97mn
---

# Fix garbled/stacked doc comments (lint-sweep fallout)

## Goal

The 2026-10-24 PR review (VERDICT: FIX_FIRST, zero blockers) and the beans
audit both flagged that the lint sweep MERGED upstream doc comments:
exported idents carry stacked duplicate blocks, some corrupted. Repair by
keeping the correct cited block and deleting the orphan duplicate — no
behavior changes, comments only.

## Research

Live garbled blocks (file:line, verified on HEAD c9fb7e8):

- internal/diablo/txtdata/classes.go:11-12 + 18-19 — ClassData/LoadClasses
  each say classdat.tsv AND classes.tsv in consecutive lines; keep ONE
  (source of truth: upstream classes.cpp / classdat table) and delete the
  other.
- internal/diablo/txtdata/spelldata.go:11-12 + 30-31 + 154-155 — same
  duplicated-intro pattern.
- internal/diablo/txtdata/mana.go:14-15 + 22-23 + 37-38 — includes literal
  corruption "mpopulas" in the port line.
- internal/diablo/missiles/process.go:23-32 + 252-262 — THREE stacked doc
  blocks on MoveMissile incl. another type's (DontMoveOnFailedProbe) doc;
  collision.go:80-86 — truncated mid-sentence.
- Sweep the rest of internal/ for the same pattern: consecutive
  doc-comment lines that restate or contradict each other (grep for
  doubled `// X is one ...` / `// X parses ...` pairs).

## Scope

- Comments only. Every kept block keeps its upstream citation
  (file:line). One doc block per ident.
- Do NOT touch: any code line, .golangci.yml, test expectations on
  behavior. gofmt must stay clean; lint must stay 0.

## Acceptance criteria

- [ ] All listed sites resolved: no stacked/contradictory doc blocks, no
      "mpopulas" or other corruption anywhere in internal/.
- [ ] `go build ./...` + full `go test -count=1 ./...` green (9696 tests,
      comments cannot break them — this proves no code lines moved).
- [ ] lint 0, gofmt clean, beans-lint OK.

## Libraries

- No new dependencies. Standard library + existing repo modules only
  (comments/JSON fixtures are data, not dependencies). Zero-dependency
  test policy (devilGo-i56j, binding 2026-10) applies.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md (comment law)
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.hermes/skills/go-dev-notes/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md


## Triage

Delegated triage evidence: brainer beans-completion audit 2026-10-24
(/tmp/devilgo-sweep/beans-audit.md, session devilGo-lu2v-brainer) verified
the gap against the tree at HEAD c9fb7e8; parent spot-checked the cited
file:line evidence before creating this bean. Scoped per audit
recommendation; acceptance criteria are checkable as written.
