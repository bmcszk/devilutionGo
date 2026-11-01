package quests

// UniqueMonsterType indices: the positions in UniqueMonstersData
// (monster.h:60-75), stored in Monster.UniqueType (monster.h:261).
const (
	// UniqueTypeGarbud is UniqueMonsterType::Garbud.
	UniqueTypeGarbud uint8 = iota
	// UniqueTypeSkeletonKing is UniqueMonsterType::SkeletonKing.
	UniqueTypeSkeletonKing
	// UniqueTypeZhar is UniqueMonsterType::Zhar.
	UniqueTypeZhar
	// UniqueTypeSnotSpill is UniqueMonsterType::SnotSpill.
	UniqueTypeSnotSpill
	// UniqueTypeLazarus is UniqueMonsterType::Lazarus.
	UniqueTypeLazarus
	// UniqueTypeRedVex is UniqueMonsterType::RedVex.
	UniqueTypeRedVex
	// UniqueTypeBlackJade is UniqueMonsterType::BlackJade.
	UniqueTypeBlackJade
	// UniqueTypeLachdan is UniqueMonsterType::Lachdan.
	UniqueTypeLachdan
	// UniqueTypeWarlordOfBlood is UniqueMonsterType::WarlordOfBlood.
	UniqueTypeWarlordOfBlood
	// UniqueTypeButcher is UniqueMonsterType::Butcher.
	UniqueTypeButcher
	// UniqueTypeHorkDemon is UniqueMonsterType::HorkDemon.
	UniqueTypeHorkDemon
	// UniqueTypeDefiler is UniqueMonsterType::Defiler.
	UniqueTypeDefiler
	// UniqueTypeNaKrul is UniqueMonsterType::NaKrul.
	UniqueTypeNaKrul
)

// UniqueTypeNone is UniqueMonsterType::None, "-1" (monster.h:74).
const UniqueTypeNone uint8 = 255

// uniqueMonsterNames is UniqueMonsterType in declaration order; the kill hook
// compares names, so the names are the enum spellings.
var uniqueMonsterNames = [...]string{
	UniqueGarbud,
	"SkeletonKing",
	UniqueZhar,
	"SnotSpill",
	UniqueLazarus,
	"RedVex",
	"BlackJade",
	"Lachdan",
	UniqueWarlordOfBlood,
	"Butcher",
	"HorkDemon",
	"Defiler",
	"NaKrul",
}

// UniqueMonsterName maps a UniqueMonsterType index to its enum name, and any
// index outside the table (including None) to the empty string.
func UniqueMonsterName(uniqueType uint8) string {
	if int(uniqueType) >= len(uniqueMonsterNames) {
		return ""
	}

	return uniqueMonsterNames[uniqueType]
}

// VictimFromMonster builds the kill hook's view of a monster from its table id
// (monster.h type().type) and its UniqueType index (monster.h:261).
func VictimFromMonster(monsterID string, uniqueType uint8) KillVictim {
	return KillVictim{MonsterID: monsterID, UniqueType: UniqueMonsterName(uniqueType)}
}
