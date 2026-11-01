package game_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game"
)

type flagCase struct {
	name        string
	args        []string
	wantEnabled bool
}

func TestNewDebugRouterFromFlag(t *testing.T) {
	// given
	cases := []flagCase{
		{name: "NoArgs", args: nil, wantEnabled: false},
		{name: "UnrelatedArgs", args: []string{"--windowed", "--verbose"}, wantEnabled: false},
		{name: "DoubleDash", args: []string{"--debug"}, wantEnabled: true},
		{name: "SingleDash", args: []string{"-debug"}, wantEnabled: true},
		{name: "AmongOtherArgs", args: []string{"--windowed", "--debug", "--verbose"}, wantEnabled: true},
	}

	// when
	routers := buildRouters(cases)

	// then
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if enabled := routers[i] != nil; enabled != tt.wantEnabled {
				t.Fatalf("NewDebugRouterFromFlag(%v) nil = %v, want enabled %v", tt.args, !enabled, tt.wantEnabled)
			}
		})
	}
}

// buildRouters resolves one router per flag case.
func buildRouters(cases []flagCase) []*game.DebugRouter {
	routers := make([]*game.DebugRouter, len(cases))
	for i, tt := range cases {
		routers[i] = game.NewDebugRouterFromFlag(tt.args, debugEnv())
	}

	return routers
}
