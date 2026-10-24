# Committed parity fixtures (devilGo-1spo)

Goldens are committed here, CI verifies hashes only — it never generates
(CI cannot fetch the copyrighted MPQs).

- `generated/` — the probe-produced fixture copy: same relative layout as
  the probe output (`dun/`, `frames/`, `lighttables/`, `rng/`, `save/`)
  plus the committed inputs (`demo_0.dmo`, `seeds.txt`) and the
  `manifest.txt` hash manifest. Regenerate with `just fixture-refresh`,
  check with `just fixture-verify`. Per-class details in each subdir's
  README.
- `static/` — hand-authored fixtures; currently empty (`.gitkeep`).
