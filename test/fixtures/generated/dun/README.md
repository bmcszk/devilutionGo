# dun/ — seeded DRLG dumps

One `.dun` + `.meta` pair per (level, entry, seed), named
`l<level>-<entry>-<seed>`: diablo L1–L4 + hellfire L17–L20, main+prev
entries, over `tools/probe/seeds.txt`. `.dun` layout: 2×u16 LE dims, the
40×40 dungeon tile array, then the 80×80 transparency window; `.meta` =
ViewPosition (2×u16 LE) + leveltype byte.

Consumers: devilGo-xcdz (dungeon parity).

Produced by `dvg_probe dun` in the vsx3 probe run committed here.
