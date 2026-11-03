# Architecture decision records

Short records for the choices that shape devilutionGo. Each ADR states the context, the decision, and what follows from it. New decisions add a new ADR instead of editing an accepted one; superseded ADRs change status, not text.

| ADR | Decision |
|-----|----------|
| [001](adr-001.md) | Go, Ebitengine, SQLite as the stack |
| [002](adr-002.md) | Fidelity mandate: 1:1 port |
| [003](adr-003.md) | SQLite-only live saves with one-time importer |
| [004](adr-004.md) | Settings stay diablo.ini, not SQLite |
| [005](adr-005.md) | MPQ assets extracted at build time |
| [006](adr-006.md) | Lua deferred, hooks stubbed now |
| [007](adr-007.md) | 640x480 logical world, integer scaling |
| [008](adr-008.md) | Three-layer DDD layout; Appendix E superseded |
| [009](adr-009.md) | Library choices pinned in a catalog |
| [010](adr-010.md) | TDD and parity gates are blocking |
| [011](adr-011.md) | Compatibility through importers, not dual formats |
| [012](adr-012.md) | SQLite driver: modernc.org/sqlite (pure Go, no CGO) |
| [013](adr-013.md) | Game tables are CSV; asset lists are YAML; settings stay INI |
| [014](adr-014.md) | One migration tool |
| [015](adr-015.md) | Deterministic fixed-tick simulation |
| [016](adr-016.md) | Networking: in-process loopback now, protocols later |
| [017](adr-017.md) | Runs without a window; graphics stays separate |
| [018](adr-018.md) | Everything is an event: commands in, events out, one ordered pass |

Sources: milestone bean [devilutionX to Go+Ebiten+SQLite whole-game conversion](../../.beans/devilGo-i56j--devilutionx-goebitensqlite-whole-game-conversion.md) (grilling decisions 2026-10-16), AGENTS.md, docs/go-libraries.md, and the triage write-backs across the bean store.