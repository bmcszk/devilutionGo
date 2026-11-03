# Architecture Decision Records

Research basis: 3 rounds, 17 sources. 3 rounds, 17 sources (Ebitengine 2.10 official release notes, DevilutionX
DeepWiki + source, Watermill docs, ark/Donburi docs, necs, quic-go,
modernc.org/sqlite, determinism references). Standing law: 1:1 **functional**
port (not code port), 100% of the game, game-agnostic framework + Diablo 1.

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-fixed-tick-deterministic-simulation-kernel.md) | Fixed-tick deterministic simulation kernel, engine-free | Proposed |
| [0002](0002-ecs-runtime-state-donburi.md) | ECS runtime state via Donburi | Proposed |
| [0003](0003-watermill-command-event-bus.md) | Command/event bus over Watermill, transport-agnostic edge | Proposed |
| [0004](0004-headless-run-modes-vm-guest-testing.md) | Headless run modes + VM-guest golden testing | Proposed |
| [0005](0005-persistence-sqlite-yaml-csv.md) | Persistence: SQLite saves, YAML configs, CSV tables | Proposed |
| [0006](0006-event-sourcing-log-snapshots-replay.md) | Event sourcing: command log, snapshots, replay | Proposed |
| [0007](0007-game-agnostic-framework.md) | Game-agnostic framework, Diablo as first content | Proposed |
