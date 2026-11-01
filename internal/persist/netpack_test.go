package persist_test

import (
	"encoding/hex"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestNetPackSizes(t *testing.T) {
	// when the pack.h:84-146 sizes are compared to the member arithmetic
	sizes := map[string][2]int{
		"ItemNetPack":   {persist.ItemNetPackSize, 2 + 2 + 4 + 18},
		"PlayerNetPack": {persist.PlayerNetPackSize, 1600 + persist.NumNetValidation*4},
	}
	// then each record matches the C++ packed-struct size
	for name, pair := range sizes {
		if pair[0] != pair[1] {
			t.Fatalf("%s size = %d, want %d", name, pair[0], pair[1])
		}
	}
}

func TestItemNetPackGoldenImage(t *testing.T) {
	// given the TItem layout of msg.h:534-547
	packed := persist.ItemNetPack{Idx: 0x0102, CreateInfo: 0x0304, Seed: 0x05060708}
	for i := range packed.Body {
		packed.Body[i] = byte(i + 1)
	}
	// when
	encoded := hex.EncodeToString(packed.Encode())
	// then the head is little-endian TItemDef and the union body follows
	const want = "02010403080706050102030405060708090a0b0c0d0e0f101112"
	if encoded != want {
		t.Fatalf("Encode = %s, want %s", encoded, want)
	}
}

func TestItemNetPackRoundTrip(t *testing.T) {
	// given a TItem image with the prepare-item tail filled
	packed := persist.PackNetItem(
		entities.Item{Itype: 3, IDidx: 75, CreateInfo: 286, Seed: 1560055601, Magical: 1, Durability: 50, MaxDur: 50},
		persist.NetItemStats{ToHit: 12, MaxDam: 34},
		persist.PackEnv{Hellfire: true},
	)
	// when
	decoded, err := persist.DecodeItemNetPack(packed.Encode())
	// then the record round-trips and keeps the prepared words
	if err != nil {
		t.Fatalf("DecodeItemNetPack: %v", err)
	}

	if decoded != packed {
		t.Fatalf("round trip = %+v, want %+v", decoded, packed)
	}

	if got := int(decoded.Body[11]) | int(decoded.Body[12])<<8; got != 12 {
		t.Fatalf("wToHit = %d, want 12", got)
	}

	if got := int(decoded.Body[13]) | int(decoded.Body[14])<<8; got != 34 {
		t.Fatalf("wMaxDam = %d, want 34", got)
	}
}

func TestPackNetItemEmpty(t *testing.T) {
	// when an empty item is packed
	packed := persist.PackNetItem(entities.Item{Itype: -1}, persist.NetItemStats{}, persist.PackEnv{})
	// then the record carries the 0xFFFF index (pack.cpp:207-210)
	if packed.Idx != -1 {
		t.Fatalf("Idx = %#x, want 0xFFFF", uint16(packed.Idx))
	}
}

func TestPlayerNetPackRoundTrip(t *testing.T) {
	// given a player with one item and the net-only extras
	player := entities.Player{NumInv: 1}
	copy(player.Name[:], "MP-Warrior")
	player.Position.Tile = engine.Point{X: 61, Y: 24}
	player.PlrLevel = 9
	player.BaseStr = 120
	player.Strength = 130
	player.Experience = 85670061
	player.Gold = 3921
	player.MemSpells = 1610612737
	player.InvList[0] = entities.Item{Itype: 3, IDidx: 82, CreateInfo: 16400, Seed: 423576018}
	player.ManaShield = true
	extras := persist.NetPlayerExtras{CharLevel: 37, FriendlyMode: true, BaseToBlk: -3}
	// when
	packed := persist.PackNetPlayer(player, extras, persist.PackEnv{Hellfire: true})
	decoded, err := persist.DecodePlayerNetPack(packed.Encode())
	// then the image is 1692 bytes and every field survives
	if err != nil {
		t.Fatalf("DecodePlayerNetPack: %v", err)
	}

	if len(packed.Encode()) != persist.PlayerNetPackSize {
		t.Fatalf("Encode len = %d, want %d", len(packed.Encode()), persist.PlayerNetPackSize)
	}

	if decoded != packed {
		t.Fatalf("round trip = %+v, want %+v", decoded, packed)
	}

	if decoded.Level != 37 || decoded.FriendlyMode != 1 || decoded.Validation[persist.ValBaseToBlk] != -3 {
		t.Fatalf("extras = %+v", decoded)
	}

	if decoded.InvList[0].Idx != 82 || decoded.Validation[persist.ValStrength] != 130 {
		t.Fatalf("items/validation = %+v", decoded)
	}
}
