package engine

import "time"

var startTime = time.Now()

// TickSource abstracts the wall clock.
type TickSource interface {
	Now() time.Duration
}

// RealTickSource reads the wall clock elapsed since process start.
type RealTickSource struct{}

func (RealTickSource) Now() time.Duration { return time.Since(startTime) }

// MockTickSource is a TickSource whose time is set by the test.
type MockTickSource struct {
	NowValue time.Duration
}

func (m *MockTickSource) Now() time.Duration { return m.NowValue }
