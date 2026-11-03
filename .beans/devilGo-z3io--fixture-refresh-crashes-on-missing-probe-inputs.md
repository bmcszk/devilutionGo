---
# devilGo-z3io
title: fixture-refresh crashes on missing probe inputs
status: draft
type: bug
priority: low
created_at: 2026-11-03T01:19:03Z
updated_at: 2026-11-03T01:19:03Z
---

## Symptom

`python3 scripts/fixture_fixtures.py refresh` aborts with a Python traceback
instead of refusing with a message, when `tools/probe/out/` exists but the
probe inputs it copies next to the artifacts (`tools/probe/demo_0.dmo`,
`tools/probe/seeds.txt`) do not:

```
  File "scripts/fixture_fixtures.py", line 62, in refresh
    shutil.copy(REPO / "tools/probe" / name, GENERATED / name)
FileNotFoundError: [Errno 2] No such file or directory:
  '.../tools/probe/demo_0.dmo'
```

Observed 2026-11-03 in the assets-wave worktree. The run also leaves a partial
`test/fixtures/generated/` tree (ignored) and deletes
`generated/manifest.txt` before it dies.

## Expected

The same shape as the existing `PROBE_OUT` guard (`scripts/fixture_fixtures.py:63`):
check both probe inputs up front and print `fixture-refresh: <path> absent — run
the probe first.`, exit nonzero, touch nothing.

## Evidence

Reproduce: have `tools/probe/out/manifest.txt` present and `tools/probe/demo_0.dmo`
absent, then run the command above.
