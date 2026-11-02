---
# devilGo-aw1c
title: 'persist: SaveItem idx remap seam'
status: completed
type: task
priority: normal
created_at: 2026-11-01T22:21:12Z
updated_at: 2026-11-02T18:54:05Z
---

# persist: SaveItem idx remap seam (PR11 brainer M1)

PR #11 brainer finding M1 (verdict /tmp/pr11-verdict.md, 2026-11-02).

`internal/persist/itemsave.go:194` writes `s.IDIdx` verbatim. Upstream
`SaveItem` (loadsave.cpp:1165-1176) computes
`idx = AllItemsList[IDidx].iMappingId` (or -1 for IDI_NONE), then for
`!gbIsHellfire && idx < IDI_NUM_DEFAULT_ITEMS` applies
`RemapItemIdxToDiablo`, for spawn `RemapItemIdxToSpawn`, and maps
-1 -> IDI_GOLD with iType=None.

The remap helpers already exist in this package (`itemremap.go`) but
`ItemSave.Encode` never calls them and has no input carrying the
mapping id (the iMappingId table lives in itemdat, not injected).

Consequence: a Go-constructed SaveItem for a Diablo/spawn save whose
entity carries a Hellfire-table index writes a wrong idx field; the
-1 -> GOLD fallback is absent. Import/export paths are unaffected
today (store blobs are opaque upstream-produced records), but the
first live write constructing SaveItem from entities diverges the .sv.

## Fix shape

Extend the encode env (mode is already passed; add injectable
`MappingIDx func(idx int32) int32` alongside `PackEnv.RemapToDiablo`)
and document the seam, or port the translation now.

## Acceptance criteria

- [ ] SaveItem.Encode applies iMappingId + RemapItemIdx per mode,
      -1 -> IDI_GOLD with iType=None
- [ ] golden test: Diablo-mode encode of a Hellfire-table idx equals
      upstream SaveItem bytes (hand-derived vector)
- [ ] recorded seam doc in itemsave.go

Parent: devilGo-6xu2 (persistence lane)
