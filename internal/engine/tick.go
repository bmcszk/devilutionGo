package engine

import "time"

// processStart is the shared origin of every tick duration, captured once at
// package init.
var processStart = time.Now()

// TickSource abstracts the wall clock so the game loop can run on a
// deterministic clock in tests and CI.
type TickSource interface {
	Now() time.Duration
}

// RealTickSource reads the wall clock elapsed since process start.
type RealTickSource struct{}

func (RealTickSource) Now() time.Duration { return time.Since(processStart) }

// MockTickSource is a TickSource whose time is set by the test.
type MockTickSource struct {
	NowValue time.Duration
}

func (m *MockTickSource) Now() time.Duration { return m.NowValue }
