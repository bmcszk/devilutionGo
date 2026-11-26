---
# devilGo-g6la
title: 'engine/assets dead code: removal decision + stale self-referential comments'
status: draft
type: task
created_at: 2026-11-26T17:56:12Z
updated_at: 2026-11-26T17:56:12Z
parent: devilGo-p1t8
---

v1on triage found internal/engine/assets has ZERO production importers (only its own test) and cannot open retail-named archives on Linux (lowercase .mpq probe). Decision owed: delete the package or record its purpose; plus stale comments - internal/assets/palette.go:32 cites *engine/assets.Stack as a satisfier (dead type), engine/assets/mpq.go:145 self-references the path it IS. Evidence: .hermes/lane/gate-parity/v1on-triage.md (corrections table + notes for the parent). Draft; triage owed.
