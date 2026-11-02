# Committed txtdata row-count goldens

Small, hand-built tables that pin the row-count contract of the txtdata loader
without copying upstream corpus data into the repo.

| Path | Header | Data rows |
|---|---|---|
| `monstdat.tsv` | 10 columns (`_monster_id … availability`), a reduced monstdat schema | **113** |
| `Experience.tsv` | `Level`, `Experience` | **50** |

Row counts are data rows only: `awk 'END{print NR}'` counts the header too, so
`monstdat.tsv` has 114 file lines and `Experience.tsv` has 51.

## Provenance

Hand-written for this repo — no upstream bytes. The schemas follow
`devilutionX/assets/txtdata/monsters/monstdat.tsv` and
`devilutionX/assets/txtdata/Experience.tsv` closely enough that the loader
exercises the same shapes (mix of id strings, enum-name strings, flags and
numbers).

The live upstream tables, when a checkout is available, are parsed by
`TestUpstreamTablesParseFully`, which fails when the loader drops or invents a
line. That test skips when no checkout is present (zero-dependency parity rule).
