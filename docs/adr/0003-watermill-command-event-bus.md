# ADR-003: Command/event bus over Watermill, transport-agnostic edge

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- Framework owns ONE `watermill.Logger`-wired `*GoChannel` pub/sub for in-process
  command/event dispatch (sync-friendly: construct with
  `gochannel.NewGoChannel(config, logger)` and keep `OutputChannelEvents` off
  unless a fire-and-forget topic is wanted — GoChannel is async by default).
- Commands from players/AI funnel into the per-tick command buffer; events out to
  presentation + (MP) network fan-out subscribers.
- Game-agnostic rule: core buses carry framework-level envelopes; game-specific
  payloads (Diablo spells etc.) are typed payloads, registered per game —
  framework never imports `game/`.
- MP transports plug in later behind one interface (LAN-first): quic-go
  (RFC 9000 + RFC 9221 unreliable datagrams, pure Go) for WAN/internet lane;
  plain UDP/TCP for LAN. SP = loopback transport, same wire format.

**Consequences.** + std-adjacent, dependency-light, backends swappable (NATS/
Redis later without touching handlers); − Watermill async semantics must be
handled deliberately per topic.
