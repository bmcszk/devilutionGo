package engine

import "time"

// nthread.cpp
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
