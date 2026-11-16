// Package inputbridge_test proves the stick-to-command bridge: the decision
// cores' outputs become typed kernel commands, nothing becomes a command when
// nothing should, and one decision yields exactly one command.
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// TestStickCommand_TestCase pins the bridge mapping: the same inputs
// ClassifyStickPush and StickDirection class on feed one command per decision,
// whose payload words carry the kind's upstream ordinals.
func TestStickCommand_TestCase(t *testing.T) {
	tests := []stickCommandCase{
		{
			name: "firm push east walks (ctrl_motion.cpp:210-218, plrctrls.cpp:1853-1861)",
			x:    1, y: 0,
			want: kernel.Command{
				Kind: kernel.KindStickPush,
				A:    int32(controls.PushFirm),
				B:    int32(controls.AxisXRight),
				C:    int32(controls.AxisYNone),
			},
		},
		{
			name: "weak push south-west turns (ctrl_motion.cpp:210-218, plrctrls.cpp:1862-1876)",
			x:    -0.4, y: -0.4,
			want: kernel.Command{
				Kind: kernel.KindStickPush,
				A:    int32(controls.PushWeak),
				B:    int32(controls.AxisXLeft),
				C:    int32(controls.AxisYDown),
			},
		},
		{
			name: "firm push south-west walks (ctrl_motion.cpp:210-218, plrctrls.cpp:1853-1861)",
			x:    -1, y: -1,
			want: kernel.Command{
				Kind: kernel.KindStickPush,
				A:    int32(controls.PushFirm),
				B:    int32(controls.AxisXLeft),
				C:    int32(controls.AxisYDown),
			},
		},
		{
			name: "released stick keeps its facing (ctrl_motion.cpp:220-242, plrctrls.cpp:1853-1876)",
			x:    0, y: 0,
			want: kernel.Command{
				Kind: kernel.KindStickPush,
				A:    int32(controls.PushNone),
				B:    int32(controls.AxisXNone),
				C:    int32(controls.AxisYDown),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := stickDecisionCommand(t, tc.x, tc.y)
			// then
			if got != tc.want {
				t.Fatalf("StickCommand = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestStickCommand_NothingToCommandDropsSilently pins the nothing decision: a
// PushNone push with no facing never becomes a command — emitting an empty
// stick command every quiet tick would flood the queue the kernel drains.
func TestStickCommand_NothingToCommandDropsSilently(t *testing.T) {
	// given: the nothing decision feeds PushNone and the NONE/NONE axis the
	// decision cores themselves produce for a dead stick.
	push := controls.PushNone
	dir := controls.AxisDirection{}

	// when
	cmd, ok := inputbridge.StickCommand(push, dir)

	// then
	if ok {
		t.Fatalf("StickCommand = %+v, want none for the nothing decision", cmd)
	}
}

// TestStickCommand_WeakAndFirmDifferOnlyInIntent pins that the bridge does not
// re-derive intent: both severities feed through the same command shape and the
// kernel handler reads the severity word.
func TestStickCommand_WeakAndFirmDifferOnlyInIntent(t *testing.T) {
	// given: a gentle and a firm push along the same pure-horizontal sector
	// (same-sign inputs, magnitudes either side of the walk threshold).
	weakX, weakY := 0.5, 0.0
	firmX, firmY := 1.0, 0.0

	// when
	weak, _ := inputbridge.StickCommand(
		controls.ClassifyStickPush(weakX, weakY), controls.StickDirection(weakX, weakY),
	)
	firm, _ := inputbridge.StickCommand(
		controls.ClassifyStickPush(firmX, firmY), controls.StickDirection(firmX, firmY),
	)

	// then
	if weak.A != int32(controls.PushWeak) || firm.A != int32(controls.PushFirm) {
		t.Fatalf("severity words weak=%d firm=%d, want the decision cores' classification verbatim",
			weak.A, firm.A)
	}
	if weak.B != firm.B || weak.C != firm.C {
		t.Fatalf("axis words differ weak=(%d,%d) firm=(%d,%d), want identical sectors for same-sign inputs",
			weak.B, weak.C, firm.B, firm.C)
	}
}
