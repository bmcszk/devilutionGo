// Package missiles_test pins the add-only cursor rows: Identify
// (missiles.cpp:2802-2818), ItemRepair (cpp:2820-2836), StaffRecharge
// (cpp:2838-2854), HealOther (cpp:2657-2672) and Resurrect (cpp:2695-2708).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// cursorAddCases pin the local-player hands: every row drops its missile and
// arms one cursor, and the identify/repair/recharge trio also closes the
// spellbook and opens the inventory.
var cursorAddCases = []cursorAddCase{
	{
		name: "identify", add: missiles.AddIdentify, cursor: enums.CursorIdentify,
		wantInventory: true, wantSpellbook: true,
	},
	{
		name: "item repair", add: missiles.AddItemRepair, cursor: enums.CursorRepair,
		wantInventory: true, wantSpellbook: true,
	},
	{
		name: "staff recharge", add: missiles.AddStaffRecharge, cursor: enums.CursorRecharge,
		wantInventory: true, wantSpellbook: true,
	},
	{name: "heal other", add: missiles.AddHealOther, cursor: enums.CursorHealOther},
	{name: "resurrect", add: missiles.AddResurrect, cursor: enums.CursorResurrect},
}

// cursorAddCase pins one cursor-handed add row.
type cursorAddCase struct {
	name          string
	add           func(*missiles.HitWorld, *missiles.Missile, *missiles.AddParameter)
	cursor        enums.CursorID
	wantInventory bool
	wantSpellbook bool
}

// cursorSpies records the UI writes of one add.
type cursorSpies struct {
	cursor     enums.CursorID
	inventory  int
	spellbooks int
}

// TestCursorAdds pins the cursor and inventory writes of the add-only rows for
// the local player, and that a remote caster writes nothing.
func TestCursorAdds(t *testing.T) {
	// given
	cases := cursorAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &cursorSpies{}
			w := cursorWorld(spies, true)
			m := missiles.Missile{Source: 0}

			// when
			tc.add(w, &m, &missiles.AddParameter{})

			// then
			assertCursorAdd(t, m, spies, tc)

			// given
			remote := &cursorSpies{}
			remoteWorld := cursorWorld(remote, false)

			// when
			tc.add(remoteWorld, &m, &missiles.AddParameter{})

			// then
			if remote.cursor != enums.CursorNone || remote.inventory != 0 {
				t.Fatalf("remote caster wrote cursor %v / inventory %d, want nothing",
					remote.cursor, remote.inventory)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// cursorWorld wires the local-player flag and the cursor/inventory writes.
func cursorWorld(spies *cursorSpies, isLocal bool) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.IsMyPlayer = func(int) bool { return isLocal }
	cb.SetCursor = func(cursor enums.CursorID) { spies.cursor = cursor }
	cb.OpenInventory = func() { spies.inventory++ }
	cb.CloseSpellbook = func() { spies.spellbooks++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// assertCursorAdd checks the handed cursor, the inventory writes and the
// dropped missile.
func assertCursorAdd(t *testing.T, m missiles.Missile, spies *cursorSpies, tc cursorAddCase) {
	t.Helper()

	if spies.cursor != tc.cursor { // cpp:2813, 2831, 2849, 2666, 2701
		t.Fatalf("cursor = %v, want %v", spies.cursor, tc.cursor)
	}

	if (spies.inventory != 0) != tc.wantInventory { // cpp:2809-2812
		t.Fatalf("inventory opens = %d, want open %v", spies.inventory, tc.wantInventory)
	}

	if (spies.spellbooks != 0) != tc.wantSpellbook { // cpp:2807-2808
		t.Fatalf("spellbook closes = %d, want close %v", spies.spellbooks, tc.wantSpellbook)
	}

	if !m.DelFlag { // cpp:2805, 2823, 2841, 2661, 2707
		t.Fatal("DelFlag = false, want the missile dropped")
	}
}
