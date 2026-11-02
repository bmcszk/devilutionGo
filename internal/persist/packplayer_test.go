package persist_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestPackPlayerFields(t *testing.T) {
	// given a player at tile (10, 20) with one inventory item
	player := entities.Player{NumInv: 1}
	player.Position.Tile = engine.Point{X: 10, Y: 20}
	copy(player.Name[:], "Conan")
	player.BaseStr = 30
	player.Experience = 1000
	player.Gold = 500
	player.HPBase = 96
	player.MaxHPBase = 128
	player.ManaBase = 32
	player.MaxManaBase = 48
	player.MemSpells = 0x0102030405060708
	player.InvList[0] = entities.Item{Itype: 3, IDidx: 5, Seed: 7}
	player.SplLvl[0] = 3
	player.SplLvl[36] = 4
	player.SplLvl[37] = 5
	player.SplLvl[46] = 6
	player.Reflections = 3
	// when
	packed := persist.PackPlayer(player, 12, persist.PackEnv{Hellfire: true})
	// then the record carries the split spell levels and the item block
	// (pack.cpp:172-216)
	if err := checkPackedPlayer(packed); err != nil {
		t.Fatal(err)
	}
}

func TestPackPlayerVanillaTargets(t *testing.T) {
	// given gbVanilla set and a player standing still
	player := entities.Player{}
	player.Position.Tile = engine.Point{X: 42, Y: 24}
	// when
	packed := persist.PackPlayer(player, 1, persist.PackEnv{Hellfire: true, Vanilla: true})
	// then the tile is copied into the target bytes (pack.cpp:162-165)
	if packed.TargX != 42 || packed.TargY != 24 {
		t.Fatalf("targets = %d,%d, want 42,24", packed.TargX, packed.TargY)
	}
}

func TestUnpackPlayerSpells(t *testing.T) {
	// given a record with every packed spell slot set
	packed := persist.PlayerPack{NumInv: 0}
	for i := range packed.SplLvl {
		packed.SplLvl[i] = uint8(i + 1)
	}

	for i := range packed.SplLvl2 {
		packed.SplLvl2[i] = uint8(i + 40)
	}

	packed.SplLvl[persist.SpellNova] = 9
	packed.SplLvl[persist.SpellApocalypse] = 11
	// when
	hellfire := persist.UnpackPlayer(packed, persist.PackEnv{Hellfire: true})
	diablo := persist.UnpackPlayer(packed, persist.PackEnv{})
	// then Hellfire keeps Nova/Apocalypse and Diablo clears them
	// (pack.cpp:403-414)
	if err := checkUnpackedSpells(hellfire, diablo); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerPackPlayerRoundTrip(t *testing.T) {
	// given a packed hero
	player := entities.Player{NumInv: 2}
	copy(player.Name[:], "Rogue")
	player.BaseDex = 40
	player.Experience = 42
	player.Gold = 9
	player.InvList[0] = entities.Item{Itype: 3, IDidx: 5, Seed: 7, Durability: 3, MaxDur: 9}
	player.InvBody[0] = entities.Item{Itype: 3, IDidx: 6, Seed: 8}
	player.SpdList[0] = entities.Item{Itype: 3, IDidx: 7, Seed: 9, Charges: 2, MaxCharges: 4}
	packed := persist.PackPlayer(player, 4, persist.PackEnv{Hellfire: true})
	// when
	got := persist.UnpackPlayer(packed, persist.PackEnv{Hellfire: true})
	// then the record-level fields come back
	if string(got.Name[:5]) != "Rogue" || got.BaseDex != 40 || got.Gold != 9 {
		t.Fatalf("player = %+v", got)
	}

	if got.InvList[0].Seed != 7 || got.InvBody[0].IDidx != 6 {
		t.Fatalf("items = %+v / %+v", got.InvList[0], got.InvBody[0])
	}

	if got.SpdList[0].Charges != 2 || got.SpdList[0].MaxCharges != 4 {
		t.Fatalf("belt item = %+v", got.SpdList[0])
	}
}

// checkPackedPlayer asserts the golden player's packed fields.
func checkPackedPlayer(packed persist.PlayerPack) error {
	if err := checkPackedStats(packed); err != nil {
		return err
	}

	return checkPackedItems(packed)
}

// checkPackedStats asserts the packed stat and spell block.
func checkPackedStats(packed persist.PlayerPack) error {
	if packed.Level != 12 || packed.Exp != 1000 || packed.Gold != 500 {
		return fmt.Errorf("stats = %+v", packed)
	}

	if packed.SplLvl[0] != 3 || packed.SplLvl[36] != 4 {
		return fmt.Errorf("SplLvl = %v", packed.SplLvl)
	}

	if packed.SplLvl2[0] != 5 || packed.SplLvl2[9] != 6 {
		return fmt.Errorf("SplLvl2 = %v", packed.SplLvl2)
	}

	return nil
}

// checkPackedItems asserts the packed item and position block.
func checkPackedItems(packed persist.PlayerPack) error {
	if packed.NumInv != 1 || packed.InvList[0].Idx != 5 {
		return fmt.Errorf("inventory = %+v", packed.InvList[0])
	}

	if packed.Px != 10 || packed.Py != 20 || packed.TargX != 0 {
		return fmt.Errorf("position = %d,%d targ %d,%d", packed.Px, packed.Py, packed.TargX, packed.TargY)
	}

	return nil
}

// checkUnpackedSpells asserts the Diablo/Hellfire spell-level difference.
func checkUnpackedSpells(hellfire, diablo entities.Player) error {
	if hellfire.SplLvl[persist.SpellNova] != 9 || hellfire.SplLvl[persist.SpellApocalypse] != 11 {
		return fmt.Errorf("hellfire spells = %v", hellfire.SplLvl)
	}

	if diablo.SplLvl[persist.SpellNova] != 0 || diablo.SplLvl[persist.SpellApocalypse] != 0 {
		return fmt.Errorf("diablo spells = %v", diablo.SplLvl)
	}

	if diablo.SplLvl[37] != 40 || diablo.SplLvl[46] != 49 {
		return fmt.Errorf("hellfire slots = %v", diablo.SplLvl[37:47])
	}

	return nil
}
