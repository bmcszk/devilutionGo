// Package testsupport: canonical state hashing for parity comparisons
// (devilGo-xcdz slice 2).
package testsupport

// HashField is one canonical state field: its name and byte encoding.
type HashField struct {
	Name  string
	Bytes []byte
}

// HashableState provides the canonical state fields in a FIXED field order.
// Canonical = player-visible state: identity, position, vitals, mode.
// NOT canonical (must be omitted): caches, derived/sprite geometry,
// map-iteration-ordered collections, timers, pointers.
type HashableState interface {
	CanonicalFields() []HashField
}

// StateHash is FNV-1a 64 over the canonical fields, serialized as
// "<name>:<hex bytes>;" per field in the provided order. Deterministic
// across runs and process restarts: no map iteration, no pointer hashing,
// no time. A caller that reorders CanonicalFields changes the hash —
// field order is part of the state contract.
func StateHash(state HashableState) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	for _, field := range state.CanonicalFields() {
		for _, b := range []byte(field.Name) {
			hash ^= uint64(b)
			hash *= prime64
		}
		hash ^= ':' //nolint:gomnd // FNV separator byte (0x3A)
		hash *= prime64
		for _, b := range field.Bytes {
			hash ^= uint64(b)
			hash *= prime64
		}
		hash ^= ';' //nolint:gomnd // field terminator byte (0x3B)
		hash *= prime64
	}
	return hash
}
