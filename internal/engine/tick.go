package engine

import "time"

// TickSource abstracts the wall clock.
type TickSource interface {
	Now() time.Duration
}

type RealTickSource struct{}

func (RealTickSource) Now() time.Duration { return time.Since(startTime) }

var startTime = time.Now()

type MockTickSource struct {
	NowValue time.Duration
}

func (m *MockTickSource) Now() time.Duration { return m.NowValue }
