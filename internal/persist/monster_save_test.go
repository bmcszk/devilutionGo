package persist_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestMonsterSaveSize(t *testing.T) {
	// when a monster record is encoded (loadsave.cpp:1508-1620)
	image := monsterFixture().Encode(persist.MonsterDerived{}, persist.MonsterWalking{})
	// then it is 216 bytes, matching the field list plus upstream's skips
	if len(image) != persist.MonsterSaveSize {
		t.Fatalf("record len = %d, want %d", len(image), persist.MonsterSaveSize)
	}
}

func TestMonsterSaveFieldOffsets(t *testing.T) {
	// when a monster record with distinct values per field is encoded (loadsave.cpp:1510-1560)
	image := monsterFixture().Encode(persist.MonsterDerived{
		Level: 7, Experience: 0x0304, ToHit: 5,
		ToHitSpecial: 6,
	}, persist.MonsterWalking{
		OffsetX: 1, OffsetY: 2, VelocityX: 3, VelocityY: 4,
		Offset2X: 5, Offset2Y: 6,
	})
	// then the head offsets match the upstream write order and skips
	cases := []struct {
		name   string
		offset int
		want   int32
	}{
		{"levelType", 0, 1},
		{"mode", 4, 2},
		{"goalVar1 (after goal u8 + 3 skip)", 12, 11},
		{"goalVar3", 20, -3},
		{"pathCount region end", 28, 4}, // tile.x starts at 32
		{"tile.x", 32, 100},
		{"old.y", 52, 109},
		{"walking offsetX (after OldY)", 56, 1},
		{"walking velocityY", 68, 4},
		{"direction", 72, 5},
		{"enemy", 76, 6},
		{"ticksPerFrame (after enemyPos + skips + ptr)", 88, 9},
		{"currentFrame+1", 100, 2},
		{"temp.y", 128, 51},
		{"offset2 deltaX", 132, 5},
		{"offset2 deltaY", 136, 6},
		{"maxHitPoints", 144, 60},
		{"hitPoints", 148, 42},
	}
	for _, tc := range cases {
		got := int32(binary.LittleEndian.Uint32(image[tc.offset : tc.offset+4]))
		if got != tc.want {
			t.Fatalf("%s at %d = %d, want %d", tc.name, tc.offset, got, tc.want)
		}
	}
}

func TestMonsterSaveDerivedAndTail(t *testing.T) {
	// when a record with derived stats and a light id of -1 is encoded (loadsave.cpp:1583-1618)
	image := monsterFixture().Encode(persist.MonsterDerived{
		Level: 7, Experience: 0x0304, ToHit: 5,
		ToHitSpecial: 6,
	}, persist.MonsterWalking{})
	// then the unique type is stored +1, the derived stats land after whoHit,
	// and NO_LIGHT becomes 0 (loadsave.cpp:1613-1617)
	cases := []struct {
		name   string
		offset int
		want   uint8
	}{
		{"uniqueType+1", 188, 3},
		{"uniqTrans", 189, 4},
		{"corpseId", 190, 0xff},
		{"whoHit", 191, 6},
		{"derived level", 192, 7},
		{"toHit", 196, 5},
		{"toHitSpecial", 199, 6},
		{"lightId NO_LIGHT", 215, 0},
	}
	for _, tc := range cases {
		if image[tc.offset] != tc.want {
			t.Fatalf("%s at %d = %d, want %d", tc.name, tc.offset, image[tc.offset], tc.want)
		}
	}

	exp := binary.LittleEndian.Uint16(image[194:196])
	if exp != 0x0304 {
		t.Fatalf("experience = %#x, want 0x0304", exp)
	}
}

func monsterFixture() persist.MonsterRecord {
	return persist.MonsterRecord{
		LevelType:        1,
		Mode:             2,
		Goal:             3,
		GoalVar1:         11,
		GoalVar2:         22,
		GoalVar3:         -3,
		PathCount:        4,
		TileX:            100,
		TileY:            110,
		FutureX:          101,
		FutureY:          111,
		OldX:             99,
		OldY:             109,
		Direction:        5,
		Enemy:            6,
		EnemyPosX:        7,
		EnemyPosY:        8,
		TicksPerFrame:    9,
		TickCounter:      10,
		NumberOfFrames:   12,
		CurrentFrame:     1,
		IsInvalid:        1,
		Var1:             -1,
		Var2:             2,
		Var3:             -3,
		TempX:            50,
		TempY:            51,
		MaxHitPoints:     60,
		HitPoints:        42,
		AI:               13,
		Intelligence:     3,
		Flags:            0xaabbccdd,
		ActiveForTicks:   5,
		LastX:            48,
		LastY:            49,
		RndItemSeed:      0x11223344,
		AiSeed:           0x55667788,
		UniqueType:       2,
		UniqTrans:        4,
		CorpseID:         -1,
		WhoHit:           6,
		MinDamage:        3,
		MaxDamage:        9,
		MinDamageSpecial: 1,
		MaxDamageSpecial: 2,
		ArmorClass:       12,
		Resistance:       0x0102,
		TalkMsg:          0x1e,
		Leader:           2,
		LeaderRelation:   1,
		PackSize:         5,
		LightID:          -1,
	}
}
