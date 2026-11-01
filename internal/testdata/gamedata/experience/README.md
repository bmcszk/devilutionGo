# Experience.tsv golden

Provenance:

- Upstream file: `assets/txtdata/Experience.tsv`
- Upstream pin: devilutionX `c20b2f2456` (READ-ONLY research checkout)
- Copied byte-identical: sha256 `c10290c4258a50794bcec0daa866c223f21f2b3cad0ff3469935060fb6705be3`
- Shape: header `Level<TAB>Experience` plus 50 data rows (levels 1-50)

The file is committed verbatim, so `internal/diablo/txtdata` can be tested
against the real ladder instead of a synthetic one. No trailing spaces, tabs
as separators, LF line endings (matching upstream).

Row-count rule: `awk 'END{print NR}' Experience.tsv` = 51 lines = 1 header +
50 data rows.
