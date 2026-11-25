---
# devilGo-lz2m
title: 'boot: per-hero single_N.db store model (ADR-003 D1 migration)'
status: todo
type: task
priority: normal
created_at: 2026-11-25T14:53:02Z
updated_at: 2026-11-25T16:27:49Z
parent: devilGo-xih1
---

# boot: per-hero single_N.db store model (ADR-003 D1 migration)

## Goal
Migrate the windowed boot's hero storage from the single multi-row single_0.db to ADR-003's per-hero model — one .db file per hero (single_N.db), the hero list probing slots 0..98 like upstream pfile_ui_set_hero_infos (pfile.cpp:695-726 @e1ad0c2b6: OpenSaveArchive(i) per slot), first-free-slot and the 99-cap enforced at file enumeration.

## Research
- hl2 triage (task-hl2-load-triage, 2026-11-25; .hermes/lane/m2-hero-load/v13s-triage-result.md section 1b): the migration is a UX + storage-handle redesign — selHeroScreen.store re-shapes from one *persist.SQLiteStore into a slot->handle map, touching refresh (selhero.go:158-176), firstFreeSlot (:467-489), performFocusedDelete (:240-254); the Load path (v13s) never touches any of it, hence the split.
- D1 decision (dcob triage block): ADR-003 "one hero = one .db file" (docs/adrs/adr-003.md:12) stands.

## Scope
- selHeroScreen storage: slot->handle map over per-hero single_N.db files.
- Hero-list enumeration + Load/New-Game flows read through the map; first-free-slot = first missing file number; 99-cap at enumeration.
- Load path (v13s) keeps working unchanged across the migration (its tests pin it).

## Acceptance criteria
- [ ] **S1 — `persist.HeroStoreSet`** (`internal/persist/hero_store_set.go`, untagged): dir+`SaveNamespace`-bound, handle-free; `Path(slot)` = `SavePath` with `db`; `SaveHero/LoadHero/ListHeroes/DeleteHero/HasSlot`; `DeleteHero` = `os.Remove`; `ListHeroes` probes 0..98 (missing = empty, corrupt = error). **RED:** `TestHeroStoreSetListHeroesProbesFiles`, `TestHeroStoreSetDeleteRemovesTheFile`, `TestHeroStoreSetPathIsGetSavePath` — `go test -count=1 -run 'TestHeroStoreSet' ./internal/persist` → `undefined: persist.NewHeroStoreSet`. Gate: `go test -count=1 ./internal/persist`; `just sqlc-check` clean.
- [ ] **S2 — `SessionArchive` routes by slot** (`session_archive.go`): `NewSessionArchive(stores *HeroStoreSet, packs)`; `SaveSlot/LoadSlot/HasGame` open the slot's own file; `world.SessionArchive` interface and kernel handlers unchanged. **RED:** `TestSessionArchiveRoutesToTheSlotsOwnFile` — `go test -count=1 -run TestSessionArchiveRoutes ./internal/persist` → constructor compile error. Gate: `go test -count=1 ./internal/persist`; `go vet ./...`.
- [ ] **S3 — boot + screen over the set** (`-tags windowed`): `openHeroStore`/`selheroStorePath`/`windowGame.store` deleted; `buildWindowGame`/`runWindowed` build `NewHeroStoreSet(dir, ns)`; selhero field + refresh/firstFreeSlot/performFocusedDelete/heroRowActivated/loadSelected/pickClass go through the set; delete = file removal (pfile.cpp:780). **RED:** `TestSelHeroListReflectsFilesOnDisk` (real temp files written directly, list reflects them), `TestSelHeroCreateWritesItsOwnFile` — `go test -count=1 -tags windowed -run 'TestSelHeroListReflectsFilesOnDisk|TestSelHeroCreateWritesItsOwnFile' ./cmd/devilutionGo` → compile error. Gate: `go test -count=1 -tags windowed ./cmd/devilutionGo`.
- [ ] **S4 — first-free by file + 99-cap + legacy re-home**: first-free = first missing file number (pfile_ui_get_first_unused_save_num, pfile.cpp:736-745); cap refused before any kernel submit (pfile.cpp:749-753); one-time sweep of a multi-row `single_0.db` → `single_<slot>.db` rows + rename to `single_0.db.pre-per-hero` (Go-only deviation, no upstream counterpart). **RED:** `TestHeroStoreSetRehomesMultiRowLegacyFile` (persist) — `go test -count=1 -run TestHeroStoreSetRehomes ./internal/persist`; cmd: the three `selhero_slots_test.go` tests green through the set. Gate: full untagged + windowed; `just lint`.
- [ ] **v13s load-path tests stay green across the migration** — every listed v13s test with mechanical edits only; no assertion weakened.
- [ ] **PoW:** per-slice RED pasted; `gofmt -l` empty; `go vet ./...` rc 0 + windowed vet; `just lint` 0 issues; `go test -count=1 ./...` ok; windowed cmd+scene without assets (AC4 env-gate excepted) and WITH assets; `just sqlc-check` clean; headless no-display pin green; manual `just run` smoke over a 2-hero legacy `single_0.db` (list shows both, marker file appears, third create lands in the first gap).
- [ ] Deviations recorded in the bean: (1) corrupt slot file = error, not skipped (upstream skips silently); (2) legacy re-home sweep is Go-only; (3) save location stays the cwd (PrefPath is a separate decision); (4) `DeleteHero` returns `ErrHeroNotFound` on a missing file.

## Libraries
stdlib + modernc.org/sqlite (existing).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/port-parity-testing/SKILL.md — the set is a ported shape (pfile_ui_set_hero_infos / pfile_delete_save) with deviation notes
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split from v13s by the hl2 brainer triage (2026-11-25, parent-verified); blocked-by v13s (the load path lands first on the single-db simplification, then this re-shapes the handle). Deep triage owed before todo: schema/handle-lifetime decisions (open-99 vs LRU handles) need a pass.

## Triage hl3 (2026-11-25, brainer task-hl3-mig-triage, parent-verified at 26943d3)

Verdict: READY. Upstream layout verified live (pfile.cpp:80-92 GetSavePath one file per saveNum; :609-612 OpenSaveArchive per call — no handle map, no LRU; :695-726 probes all 99 slots; :736-753 first-free + 99-cap; :775-783 delete = file removal). Architecture: `persist.HeroStoreSet` — dir+namespace-bound, handle-free, open-on-demand-then-close (persist already owns SavePath/HeroSlots law, pfile.go:59-72; the migrate CLI already names files single_N.db). The surprise: the REAL redesign is the kernel-registered SessionArchive (bound to ONE *SQLiteStore at boot) learning to route by slot — slice S2. The selhero screen rewiring is mechanical (4 call sites). Deviations (1)-(4) recorded above. Full triage: .hermes/lane/m2-hero-load/lz2-triage-result.md (gitignored).
