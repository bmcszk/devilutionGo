package game_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// levelWarpCase is one cheat line and the level its command must carry.
type levelWarpCase struct {
	name  string
	line  string
	level int32
}

func TestDebugCommand_LevelWarpMapsToACommand(t *testing.T) {
	// given
	cases := []levelWarpCase{
		{name: "town", line: "jump-level 1", level: 1},
		{name: "hell", line: "jump-level 16", level: 16},
		{name: "surrounding spaces", line: "  jump-level   7 ", level: 7},
	}

	// when
	commands, mapped := mapLevelWarps(cases)

	// then
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := commands[i]

			if !mapped[i] || cmd.Kind != kernel.KindDebugJumpLevel || cmd.A != tc.level {
				t.Fatalf("DebugCommand(%q) = {%q %d} mapped=%v, want {%q %d}", tc.line, cmd.Kind, cmd.A,
					mapped[i], kernel.KindDebugJumpLevel, tc.level)
			}
		})
	}
}

func TestDebugCommand_LeavesEveryOtherLineUnmapped(t *testing.T) {
	// given
	lines := []string{
		"",
		"spawn-monster zombie 3",
		"spawn-item short sword",
		"teleport 4",
		"jump-level 0",
		"jump-level deep",
		"jump-level 4 5",
	}

	// when
	raised := make([]bool, len(lines))

	for i, line := range lines {
		_, raised[i] = game.DebugCommand(line)
	}

	// then
	for i, line := range lines {
		if raised[i] {
			t.Fatalf("DebugCommand(%q) raised a command, want none on the integer path", line)
		}
	}
}

// mapLevelWarps maps every line through the debug command producer.
func mapLevelWarps(cases []levelWarpCase) ([]kernel.Command, []bool) {
	commands := make([]kernel.Command, len(cases))
	mapped := make([]bool, len(cases))

	for i, tc := range cases {
		commands[i], mapped[i] = game.DebugCommand(tc.line)
	}

	return commands, mapped
}
