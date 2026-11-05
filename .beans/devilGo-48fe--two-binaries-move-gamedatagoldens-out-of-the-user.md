---
# devilGo-48fe
title: 'Two binaries: move gamedatagoldens out of the user-facing command set'
status: todo
type: task
priority: normal
created_at: 2026-11-05T22:27:28Z
updated_at: 2026-11-05T22:27:42Z
parent: devilGo-8csd
---

# Two binaries: move gamedatagoldens out of the user-facing command set

## Goal
`cmd/` holds only user-facing binaries. `gamedatagoldens` is a fixture
refresher under the ADR-019 carve-out, so it leaves `cmd/` for the devkit and
test-instrumentation surface. Recording decision 2 of docs/adrs/adr-020.md.

## Research (pre-validated)
- ADR-020 classifies `cmd/gamedatagoldens` as test instrumentation, not a
  conversion binary (docs/adrs/adr-020.md:46-58).
- It regenerates committed golden JSON from committed TSVs and never touches
  devilutionX, an MPQ or the network (cmd/gamedatagoldens/main.go:22,:26,
  tools/probe/README.md:117-122).
- ADR-019 already calls it a fixture refresher next to `tools/probe/dunconv`
  and `tools/probe/affix_probe` (docs/adrs/adr-019.md:81-84), and the devkit
  table in tools/probe/README.md:96 already lists it.
- Call sites: `justfile:57-66` (`refresh-gamedata-goldens` runs
  `go run ./cmd/gamedatagoldens` at `justfile:64`), tools/probe/README.md:96,:117-125,
  docs/research/dg-conversion-inventory.md:25,:41,:121.
  Consumers are the `internal/diablo/gamedata/audit` tests, which read the
  committed JSON, not the generator.

## Scope
- Decide the final home when executing: `tools/goldens` (one devkit command
  next to `dunconv` and `affix_probe`) or deletion if a justfile-only recipe
  with `go run` cannot justify a package. The ADR leaves both open.
- Move or delete `cmd/gamedatagoldens/`; keep the generator behaviour and the
  output path `internal/testdata/gamedata/goldens` byte-identical
  (cmd/gamedatagoldens/main.go:22).
- Retarget `justfile:57-66` and the three docs references above.
- Do not touch `internal/testdata/gamedata/**` or any golden.

## Acceptance criteria
- [ ] `devilutionGo` and `devilutionGo-migrate` are the only command dirs under `cmd/`
- [ ] `just refresh-gamedata-goldens` still regenerates the goldens from committed TSVs and reports a clean tree
- [ ] `git diff` on `internal/testdata/gamedata/goldens` is empty after one run
- [ ] no live reference to `cmd/gamedatagoldens` outside ADR-019/020 and history files
- [ ] gofmt, vet, golangci-lint, `go test ./... -count=1`, beans-lint all clean

## Triage
Triage evidence: no separate brainer session ran for this bean; this lane's
own disk-first read is the source of record and it landed as ADR-020, written
from the real tree (docs/adrs/adr-020.md).
Reading `cmd/gamedatagoldens/main.go` and `justfile:57-66` shows the generator
has no runtime dependency on the game and no test imports it, so a move cannot
change game behaviour. The only risk is a stale recipe path, which the
acceptance criteria check by running the recipe.

## Libraries
- none (file move plus recipe retarget; stdlib only)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/beans/SKILL.md` — read fully BEFORE acting
