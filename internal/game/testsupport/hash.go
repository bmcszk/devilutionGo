package testsupport

// HashField is one canonical state field: its name and byte encoding.
type HashField struct {
	Name  string
	Bytes []byte
}

// HashableState provides the canonical state fields in a fixed field order.
// Canonical = player-visible state: identity, position, vitals, mode.
// Must be omitted: caches, derived/sprite geometry, map-iteration-ordered
// collections, timers, pointers.
type HashableState interface {
	CanonicalFields() []HashField
}

// StateHash is FNV-1a 64 over the canonical fields, serialized as
// "<name>:<bytes>;" per field in the provided order.
// Deterministic across runs: no map iteration, no pointer hashing, no time.
// Field order is part of the state contract — reordering CanonicalFields
// changes the hash.
func StateHash(state HashableState) uint64 {
	const (
		offset64   = 14695981039346656037
		prime64    = 1099511628211
		separator  = ':'
		terminator = ';'
	)

	hash := uint64(offset64)

	for _, field := range state.CanonicalFields() {
		for _, b := range []byte(field.Name) {
			hash ^= uint64(b)
			hash *= prime64
		}

		hash ^= uint64(separator)

		hash *= prime64
		for _, b := range field.Bytes {
			hash ^= uint64(b)
			hash *= prime64
		}

		hash ^= uint64(terminator)
		hash *= prime64
	}

	return hash
}
