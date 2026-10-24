package engine

import "time"

// TickSource abstracts the wall clock so the game loop can run on a
// deterministic clock in tests and CI.
type TickSource interface {
	Now() time.Duration
}

// MockTickSource is a TickSource whose time is set by the test.
type MockTickSource struct {
	NowValue time.Duration
}

func (m *MockTickSource) Now() time.Duration { return m.NowValue }
