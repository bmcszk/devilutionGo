// Package levels ports the Diablo dungeon generators: the cathedral levels
// (levels/drlg_l1.cpp, DTYPE_CATHEDRAL), the Hellfire crypt variant
// (levels/crypt.cpp, DTYPE_CRYPT) and the general dungeon helpers they share
// (levels/gendung.cpp).
//
// Upstream these generators mutate file-scope globals (levels/dun_tile_data.cpp:
// 31-62). Here the same arrays live in one [Dungeon] value: one instance is one
// set of those globals, which is what lets a test script a level without any
// package-level mutable state.
//
// Two tile vocabularies exist. Cathedral and crypt tile ids are separate C++
// enums living in separate anonymous namespaces, and the two disagree on
// values for the same name (VWall5 is 90 in the cathedral, VWall6 is 90 in the
// crypt). Go has one namespace per package, so the constant sets are prefixed
// cat* / crypt*. The numeric ids are the contract; the prefixes only keep two
// upstream vocabularies apart.
package levels
