# lighttables/ — lighting engine tables

Native contiguous C++ arrays from all four `MakeLightTable()` leveltype
branches: `cathedral_tables.bin` / `hell_tables.bin` / `nest_tables.bin` /
`crypt_tables.bin` (16×256 table bytes each), `falloffs.bin` (16×128) and
`cone.bin` (8×8×16×16).

Consumers: lighting parity tests; devilGo-kx0l (falloffs/cone goldens).

Produced by `dvg_probe lighttables` in the vsx3 probe run committed here.
