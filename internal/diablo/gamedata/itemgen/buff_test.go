package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestSetupAllItemsHellfireBuff(t *testing.T) {
	tests := []struct {
		name     string
		hellfire bool
		want     int
	}{
		{name: "hellfire sets the buff bit", hellfire: true, want: itemgen.BuffHellfire},
		{name: "vanilla leaves the buff clear", hellfire: false, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			items := []itemdat.ItemData{{ID: "SWORD", Name: "Sword", ItemType: "Sword", Durability: 100}}
			req := itemgen.SetupRequest{
				ItemIndex: 0, Seed: 1, Level: 10, Uper: 1,
				Env: itemgen.SpellEnv{Hellfire: tc.hellfire},
			}
			rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}

			// when
			plan, err := itemgen.SetupAllItems(items, nil, nil, nil, req, rnd.rnd, func(int) {})

			// then
			assertPlanBuff(t, plan, err, tc.want)
		})
	}
}

func TestUidOffsetBits(t *testing.T) {
	tests := []struct {
		name      string
		uidOffset int
		want      int
	}{
		{name: "zero", uidOffset: 0, want: 0},
		{name: "one", uidOffset: 1, want: 2},
		{name: "eight", uidOffset: 8, want: 16},
		{name: "fifteen", uidOffset: 15, want: 30},
		{name: "sixteen is masked out", uidOffset: 16, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			bits := itemgen.UidOffsetBits(tc.uidOffset)

			// then
			if bits != tc.want {
				t.Fatalf("bits = %#x, want %#x", bits, tc.want)
			}
		})
	}
}

func assertPlanBuff(t *testing.T, plan itemgen.SetupPlan, err error, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Buff != want {
		t.Fatalf("buff = %#x, want %#x", plan.Buff, want)
	}
}
