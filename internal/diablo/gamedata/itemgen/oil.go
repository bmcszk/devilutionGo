package itemgen

import "fmt"

// Gold limits and cursors (items.h:29-31, tables/itemdat.h:116-118).
const (
	GoldSmallLimit  = 1000
	GoldMediumLimit = 2500

	CursorGoldSmall  = 4
	CursorGoldMedium = 5
	CursorGoldLarge  = 6
)

// oilCount is the OilLevels/OilNames table length (items.cpp:202-228).
const oilCount = 10

// multiplayerOilIndexes are the two oils the multiplayer branch allows
// (items.cpp:1290-1294): Blacksmith Oil and Oil of Imperviousness.
var multiplayerOilIndexes = [...]int{5, 9}

// oilLevels, oilValues and oilNames are the Hellfire oil tables
// (items.cpp:202-228).
var (
	oilLevels = [oilCount]int{1, 10, 1, 10, 4, 1, 5, 17, 1, 10}
	oilValues = [oilCount]int{500, 2500, 500, 2500, 1500, 100, 2500, 15000, 500, 2500}
	oilNames  = [oilCount]string{
		"Oil of Accuracy", "Oil of Mastery", "Oil of Sharpness", "Oil of Death",
		"Oil of Skill", "Blacksmith Oil", "Oil of Fortitude", "Oil of Permanence",
		"Oil of Hardening", "Oil of Imperviousness",
	}
	oilMagic = [oilCount]string{
		"OILACC", "OILMAST", "OILSHARP", "OILDEATH", "OILSKILL",
		"OILBSMTH", "OILFORT", "OILPERM", "OILHARD", "OILIMP",
	}
)

// Oil is the GetOilType outcome (items.cpp:1286-1308).
type Oil struct {
	Index int
	Name  string
	Misc  string
	Value int
}

// GetOilType ports GetOilType (items.cpp:1286-1308): build the level-filtered
// oil pool, then draw one and take its name, misc id and value.
func GetOilType(env SpellEnv, maxLvl int, rnd func(int32) int32) (Oil, error) {
	pool, err := oilPool(env, maxLvl)
	if err != nil {
		return Oil{}, err
	}

	index := pool[int(rnd(int32(len(pool))))]
	if index < 0 || index >= oilCount {
		return Oil{}, fmt.Errorf("oil index %d out of range", index)
	}

	return Oil{Index: index, Name: oilNames[index], Misc: oilMagic[index], Value: oilValues[index]}, nil
}

// GetGoldCursor is GetGoldCursor (items.cpp:2394-2401).
func GetGoldCursor(value int) int {
	if value >= GoldMediumLimit {
		return CursorGoldLarge
	}
	if value <= GoldSmallLimit {
		return CursorGoldSmall
	}

	return CursorGoldMedium
}

// oilPool is the `rnd` initialiser plus the level filter (:1289-1297).
func oilPool(env SpellEnv, maxLvl int) ([]int, error) {
	if env.Multiplayer {
		return multiplayerOilIndexes[:], nil
	}
	if maxLvl == 0 {
		maxLvl = 1
	}

	pool := make([]int, 0, oilCount)
	for index, level := range oilLevels {
		if level <= maxLvl {
			pool = append(pool, index)
		}
	}
	if len(pool) == 0 {
		return nil, fmt.Errorf("no oil available at level %d", maxLvl)
	}

	return pool, nil
}
