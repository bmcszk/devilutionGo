package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestSetupAllUsefulVanilla(t *testing.T) {
	tests := []struct {
		name  string
		level int
		draws []int32
		want  string
	}{
		{name: "mana on draw 0", level: 1, draws: []int32{0}, want: "IDI_MANA"},
		{name: "heal on draw 1", level: 1, draws: []int32{1}, want: "IDI_HEAL"},
		// level > 1: FlipCoin(3) promotes to a town portal on draw 0.
		{name: "portal on level 2", level: 2, draws: []int32{1, 0}, want: "IDI_PORTAL"},
		{name: "keeps the potion when the coin fails", level: 2, draws: []int32{1, 1}, want: "IDI_HEAL"},
		{name: "no portal roll at level 1", level: 1, draws: []int32{1}, want: "IDI_HEAL"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			rnd := &tapeRnd{t: t, values: tc.draws}

			// when
			plan, err := itemgen.SetupAllUseful(itemgen.UsefulRequest{Level: tc.level}, rnd.rnd)

			// then
			assertUseful(t, plan, err, tc.want, tc.level)
		})
	}
}

func TestSetupAllUsefulHellfire(t *testing.T) {
	tests := []struct {
		name  string
		level int
		draw  int32
		want  string
	}{
		{name: "portal above level 1", level: 2, draw: 0, want: "IDI_PORTAL"},
		{name: "heal at level 1", level: 1, draw: 0, want: "IDI_HEAL"},
		{name: "heal on 1", level: 5, draw: 1, want: "IDI_HEAL"},
		{name: "heal on 2", level: 5, draw: 2, want: "IDI_HEAL"},
		{name: "mana at level 1", level: 1, draw: 3, want: "IDI_MANA"},
		{name: "portal on 3 above level 1", level: 9, draw: 3, want: "IDI_PORTAL"},
		{name: "mana on 4", level: 5, draw: 4, want: "IDI_MANA"},
		{name: "mana on 5", level: 5, draw: 5, want: "IDI_MANA"},
		{name: "oil on 6", level: 5, draw: 6, want: "IDI_OIL"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			rnd := &tapeRnd{t: t, values: []int32{tc.draw}}

			// when
			plan, err := itemgen.SetupAllUseful(itemgen.UsefulRequest{Level: tc.level, Hellfire: true}, rnd.rnd)

			// then
			assertUseful(t, plan, err, tc.want, tc.level)
			if rnd.args[0] != 7 {
				t.Fatalf("draw arg = %d, want the Hellfire 1d7", rnd.args[0])
			}
		})
	}
}

func assertUseful(t *testing.T, plan itemgen.UsefulPlan, err error, want string, level int) {
	t.Helper()

	if err != nil {
		t.Fatalf("SetupAllUseful: %v", err)
	}
	if plan.Item != want {
		t.Fatalf("item = %q, want %q", plan.Item, want)
	}
	if plan.CreateInfo != level|itemgen.CreateInfoUseful {
		t.Fatalf("create info = %#x, want %#x", plan.CreateInfo, level|itemgen.CreateInfoUseful)
	}
}
