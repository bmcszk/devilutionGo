package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// TestDamageWeaponBreaksLeftWeapon pins the break branch: the left weapon loses
// a point and a zero durability removes it (player.cpp:478-490).
func TestDamageWeaponBreaksLeftWeapon(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 1,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 0, coinZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	assertBroken(t, out, itemgen.HandLeft)
}

// TestDamageWeaponIndestructible pins the DUR_INDESTRUCTIBLE guard
// (player.cpp:479-481, 490-492).
func TestDamageWeaponIndestructible(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: itemgen.DurIndestructible,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 0, coinZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 0 {
		t.Fatalf("actions = %+v, want none for an indestructible weapon", out.Actions)
	}
}

// TestDamageWeaponDecayRemovesAtMinus100 pins WeaponDecay's removal threshold
// (player.cpp:447-453).
func TestDamageWeaponDecayRemovesAtMinus100(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 5, PLDam: -98, Decay: true,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 0, coinAboveZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 1 || out.Actions[0].Kind != itemgen.WearBroken || out.Actions[0].PLDam != -103 {
		t.Fatalf("actions = %+v, want the decay removal at -103", out.Actions)
	}
}

// TestDamageWeaponDecayThenDurability pins the two-step pass: decay applies
// first and the durability hit still lands afterwards (player.cpp:469-487).
func TestDamageWeaponDecayThenDurability(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 5, Decay: true,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 4, coinZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 2 {
		t.Fatalf("actions = %+v, want decay then damage", out.Actions)
	}
	if out.Actions[0].Kind != itemgen.WearDecayed || out.Actions[0].PLDam != -5 {
		t.Fatalf("first action = %+v, want the -5 decay", out.Actions[0])
	}
	if out.Actions[1].Kind != itemgen.WearDamaged || out.Actions[1].Durability != 4 {
		t.Fatalf("second action = %+v, want durability 4", out.Actions[1])
	}
}

// TestDamageWeaponCoinFails pins the FlipCoin(damageFrequency) gate: a non-zero
// draw ends the pass (player.cpp:471-473).
func TestDamageWeaponCoinFails(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 5,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 4, coinAboveZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 0 {
		t.Fatalf("actions = %+v, want none when the coin fails", out.Actions)
	}
}

// TestDamageWeaponRightWeaponDamaged pins the right-hand block
// (player.cpp:488-499).
func TestDamageWeaponRightWeaponDamaged(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{Empty: true}, Right: itemgen.HandState{
		ItemType: "Axe", Class: "Weapon", Durability: 5,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 4, coinZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 1 || out.Actions[0].Slot != itemgen.HandRight || out.Actions[0].Durability != 4 {
		t.Fatalf("actions = %+v, want the right weapon at durability 4", out.Actions)
	}
}

// TestDamageWeaponShieldInFreeHand pins the shield blocks, which only fire when
// the other hand is empty (player.cpp:501-521).
func TestDamageWeaponShieldInFreeHand(t *testing.T) {
	// given
	state := itemgen.WeaponState{Left: itemgen.HandState{Empty: true}, Right: itemgen.HandState{
		ItemType: "Shield", Class: "Armor", Durability: 3,
	}}

	// when
	out, err := itemgen.DamageWeapon(state, 4, coinZero)
	// then
	if err != nil {
		t.Fatalf("DamageWeapon: %v", err)
	}
	if len(out.Actions) != 1 || out.Actions[0].Kind != itemgen.WearDamaged || out.Actions[0].Durability != 2 {
		t.Fatalf("actions = %+v, want the shield at durability 2", out.Actions)
	}
}

// TestDamageWeaponBreakCallsCalcPlrInv pins the AC's CalcPlrInv requirement: a
// break removes the equipment and re-derives the player stats
// (player.cpp:482-483, 494-495, 505-506, 516-517), while a plain durability
// tick does not (:485-486).
func TestDamageWeaponBreakCallsCalcPlrInv(t *testing.T) {
	// given
	broken := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 1,
	}}
	damaged := itemgen.WeaponState{Left: itemgen.HandState{
		ItemType: "Sword", Class: "Weapon", Durability: 4,
	}}

	// when
	breakOut, errBreak := itemgen.DamageWeapon(broken, 0, coinZero)
	tickOut, errTick := itemgen.DamageWeapon(damaged, 0, coinZero)
	// then
	if errBreak != nil || errTick != nil {
		t.Fatalf("DamageWeapon: %v / %v", errBreak, errTick)
	}
	if !breakOut.Actions[0].CalcPlrInv {
		t.Fatalf("action = %+v, want CalcPlrInv on the break", breakOut.Actions[0])
	}
	if tickOut.Actions[0].CalcPlrInv {
		t.Fatalf("action = %+v, want no CalcPlrInv for a surviving tick", tickOut.Actions[0])
	}
}

func coinZero(int32) int32 {
	return 0
}

func assertBroken(t *testing.T, out itemgen.WearResult, slot itemgen.HandSlot) {
	t.Helper()

	if len(out.Actions) != 1 {
		t.Fatalf("actions = %+v, want exactly the break", out.Actions)
	}
	if got := out.Actions[0]; got.Slot != slot || got.Kind != itemgen.WearBroken || got.Durability > 0 {
		t.Fatalf("action = %+v, want a break in the %s hand", got, slot)
	}
}

func coinAboveZero(int32) int32 {
	return 1
}
