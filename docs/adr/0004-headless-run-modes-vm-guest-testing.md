# ADR-004: Headless run modes + VM-guest golden testing

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- Three run modes from one binary: `sim` (headless sim, tick loop + checksums),
  `server` (headless MP host), `client` (Ebiten presentation). Modes = main.go
  wiring differences only; identical kernel + commands.
- Headless coverage in CI via Ebitengine 2.10 **Application Virtualization**:
  ebitenginevmguest build tag, host drives ticks/input, captures render/audio
  frames for golden tests. Also plain `go test ./sim/...` (no window needed
  since kernel is engine-free).
- Audio/render behind engine-owned interfaces from day one (`Renderer`, `Mixer`
  interfaces in framework; Ebiten implements them; future engine swap = new
  implementations, zero kernel change).

**Consequences.** + CI can play the game without a GPU/window; AI agents can
drive builds headlessly; − slight indirection cost on hot paths (measured, not
assumed).
