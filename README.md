# devilutionGo

devilutionGo is a free, open-source reimplementation of **Diablo 1** (and
Hellfire) built on [devilutionX](https://github.com/diasurgical/devilutionX) —
rewritten from C++ to **Go** so it runs anywhere Go runs, with the same look,
feel, and behavior as the original game.

**Status: early development.** The engine is not yet playable. See
[Roadmap](#roadmap).

## What is this?

- Plays the original Diablo / Hellfire game data (you must own the game)
- Same algorithms, same monsters, same loot — a 1:1 port, not a remake
- Cross-platform via [Ebitengine](https://ebitengine.org) (Windows, macOS, Linux)
- Modern save system (SQLite), still compatible with original save files

## Install (players)

No releases yet. Once the first playable build lands:

1. Install [Go 1.27+](https://go.dev/dl/) and [just](https://github.com/casey/just)
2. Get the source:

   ```
   git clone https://github.com/bmcszk/devilutionGo
   cd devilutionGo
   just build-gui
   ```

3. Run `bin/devilutionGo`

You will need the original DIABDAT.MPQ from your own Diablo 1 (or Hellfire)
copy — placed where the game asks on first launch.

## Roadmap

- **M0** — engine scaffold, CI, parity harness *(in progress)*
- **M1** — playable Spawn demo: town + cathedral
- **M2** — full single-player Diablo + Hellfire
- **M3** — multiplayer
- **M4** — polish, releases

Full architecture and development workflow: [AGENTS.md](AGENTS.md)
(dev conventions: [DEVELOPMENT.md](DEVELOPMENT.md)).
