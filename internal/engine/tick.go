package engine

import "time"

// TickSource provides the engine's notion of elapsed time so the tick loop can
// be driven by the wall clock in production and a mock in tests (see
// devilutionX nthread.cpp tick accounting). It must be safe for the single
// goroutine that owns the game loop.
type TickSource interface {
	Now() time.Duration
}

// RealTickSource reads the monotonic wall clock. It is the production
// TickSource and holds no state, so one shared value is safe to reuse.
type RealTickSource struct{}

// Now returns time.Since of the process-start monotonic base.
func (RealTickSource) Now() time.Duration { return time.Since(startTime) }

var startTime = time.Now()

// MockTickSource is a TickSource whose clock only moves when the test moves
// it, making tick accounting fully deterministic.
type MockTickSource struct {
	NowValue time.Duration
}

// Now returns the mutable NowValue.
func (m *MockTickSource) Now() time.Duration { return m.NowValue }
