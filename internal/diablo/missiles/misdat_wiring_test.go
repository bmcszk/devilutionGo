// Package missiles_test pins the wired rows against the committed misdat
// corpus: the addFn/processFn columns must name the ported bodies, and the
// row's graphic must produce the frame group AddMissile picks
// (missiles.cpp:2834-2837, animFAmt = the sprite's numFrames, misdat.cpp:152).
package missiles_test

import (
	"os"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

const (
	misdatPath  = "../../../internal/testdata/gamedata/misdat/misdat.tsv"
	spritesPath = "../../../internal/testdata/gamedata/misdat/missile_sprites.tsv"
)

// wiredRows is the slice-1 scope in misdat column terms.
var wiredRows = []wiredRow{
	{missiles.Arrow, "Arrow", "AddArrow", "ProcessArrow", "Arrow"},
	{missiles.Firebolt, "Firebolt", "AddFirebolt", "ProcessGenericProjectile", "Fireball"},
	{missiles.MagmaBall, "MagmaBall", "AddMagmaBall", "ProcessGenericProjectile", "MagmaBall"},
	{missiles.BloodStar, "BloodStar", "AddGenericMagicMissile", "ProcessGenericProjectile", "BloodStar"},
	{missiles.Acid, "Acid", "AddAcid", "ProcessGenericProjectile", "Acid"},
	{missiles.Fireball, "Fireball", "AddFireball", "ProcessFireball", "Fireball"},
	{missiles.NovaBall, "NovaBall", "AddNovaBall", "ProcessNovaBall", "Lightning"},
	{
		missiles.MagmaBallExplosion, "MagmaBallExplosion", "AddMissileExplosion",
		"ProcessMissileExplosion", "MagmaBallExplosion",
	},
	{missiles.ChargedBolt, "ChargedBolt", "AddChargedBolt", "ProcessChargedBolt", "ChargedBolt"},
	{missiles.HolyBolt, "HolyBolt", "AddHolyBolt", "ProcessHolyBolt", "HolyBolt"},
	{missiles.Lightning, "Lightning", "AddLightning", "ProcessLightning", "Lightning"},
	{missiles.ThinLightning, "ThinLightning", "AddLightning", "ProcessLightning", "ThinLightning"},
	{missiles.FireWall, "FireWall", "AddFireWall", "ProcessFireWall", "FireWall"},
	{missiles.FlameWave, "FlameWave", "AddFlameWave", "ProcessFlameWave", "FireWall"},
	{missiles.FlashTop, "FlashTop", "AddFlashTop", "ProcessFlashTop", "FlashTop"},
	{missiles.FlashBottom, "FlashBottom", "AddFlashBottom", "ProcessFlashBottom", "FlashBottom"},
	{missiles.ResurrectBeam, "ResurrectBeam", "AddResurrectBeam", "ProcessResurrectBeam", "Resurrect"},
	{missiles.AcidSplat, "AcidSplat", "AddMissileExplosion", "ProcessAcidSplate", "AcidSplat"},
	{missiles.Nova, "Nova", "AddNova", "ProcessNova", "Lightning"},
	{missiles.RedPortal, "RedPortal", "AddRedPortal", "ProcessRedPortal", "RedPortal"},
	{missiles.AcidPuddle, "AcidPuddle", "AddAcidPuddle", "ProcessAcidPuddle", "AcidPuddle"},
	{missiles.Inferno, "Inferno", "AddInferno", "ProcessInferno", "Inferno"},
	// An invisible row draws no sprite, so its graphic column is empty
	// (misdat.tsv:41).
	{missiles.Infravision, "Infravision", "AddInfravision", "ProcessInfravision", ""},
	{missiles.ChainLightning, "ChainLightning", "AddChainLightning", "ProcessChainLightning", "Lightning"},
	{missiles.LightningControl, "LightningControl", "AddLightningControl", "ProcessLightningControl", "Lightning"},
	{missiles.WeaponExplosion, "WeaponExplosion", "AddWeaponExplosion", "ProcessWeaponExplosion", ""},
	{missiles.InfernoControl, "InfernoControl", "AddInfernoControl", "ProcessInfernoControl", ""},
	{
		missiles.ThinLightningControl,
		"ThinLightningControl",
		"AddLightningControl",
		"ProcessLightningControl",
		"ThinLightning",
	},
	{missiles.ManaShield, "ManaShield", "AddManaShield", "", "ManaShield"},
	{missiles.Healing, "Healing", "AddHealing", "", ""},
	{missiles.Rage, "Rage", "AddRage", "ProcessRage", ""},
	{missiles.Guardian, "Guardian", "AddGuardian", "ProcessGuardian", "Guardian"},
	{missiles.Apocalypse, "Apocalypse", "AddApocalypse", "ProcessApocalypse", "ApocalypseBoom"},
	{missiles.Elemental, "Elemental", "AddElemental", "ProcessElemental", "Elemental"},
	{missiles.BoneSpirit, "BoneSpirit", "AddBoneSpirit", "ProcessBoneSpirit", "BoneSpirit"},
	{missiles.Golem, "Golem", "AddGolem", "", ""},
	{missiles.FireWallControl, "FireWallControl", "AddWallControl", "ProcessWallControl", "FireWall"},
	{missiles.FlameWaveControl, "FlameWaveControl", "AddFlameWaveControl", "ProcessFlameWaveControl", "FireWall"},
	{missiles.DiabloApocalypse, "DiabloApocalypse", "AddDiabloApocalypse", "", ""},
	{missiles.Identify, "Identify", "AddIdentify", "", ""},
	{missiles.ItemRepair, "ItemRepair", "AddItemRepair", "", ""},
	{missiles.StaffRecharge, "StaffRecharge", "AddStaffRecharge", "", ""},
	{missiles.HealOther, "HealOther", "AddHealOther", "", ""},
	{missiles.Resurrect, "Resurrect", "AddResurrect", "", ""},
	{missiles.TrapDisarm, "TrapDisarm", "AddTrapDisarm", "", ""},
	{missiles.Telekinesis, "Telekinesis", "AddTelekinesis", "", ""},
	{missiles.ChainBall, "ChainBall", "", "", "Lightning"},
	{missiles.BloodHit, "BloodHit", "", "", "BloodHit"},
	{missiles.BoneHit, "BoneHit", "", "", "BoneHit"},
	{missiles.MetalHit, "MetalHit", "", "", "MetalHit"},
	{missiles.DoomSerpents, "DoomSerpents", "", "", "DoomSerpents"},
	{missiles.FireOnly, "FireOnly", "", "", "FireWall"},
	{missiles.BloodRitual, "BloodRitual", "", "", ""},
	{missiles.Invisibility, "Invisibility", "", "", ""},
	{missiles.Etherealize, "Etherealize", "", "", "Etherealize"},
	{missiles.Spurt, "Spurt", "", "", "Spurt"},
	{missiles.FireMan, "FireMan", "", "", ""},
	{missiles.Krull, "Krull", "", "", "Krull"},
	{
		missiles.ApocalypseBoom, "ApocalypseBoom", "AddApocalypseBoom",
		"ProcessApocalypseBoom", "ApocalypseBoom",
	},
	{
		missiles.DiabloApocalypseBoom, "DiabloApocalypseBoom", "AddApocalypseBoom",
		"ProcessApocalypseBoom", "DiabloApocalypseBoom",
	},
	{
		missiles.BloodStarExplosion, "BloodStarExplosion", "AddMissileExplosion",
		"ProcessMissileExplosion", "BloodStarExplosion",
	},
	{missiles.FireArrow, "FireArrow", "AddElementalArrow", "ProcessElementalArrow", "FireArrow"},
	{
		missiles.LightningArrow, "LightningArrow", "AddElementalArrow",
		"ProcessElementalArrow", "LightningArrow",
	},
}

// wiredRow pins one misdat row against the port.
type wiredRow struct {
	id        missiles.MissileID
	name      string
	addFn     string
	processFn string
	graphic   string
}

// TestMisdatRowsMatchWiredBodies: every wired id's corpus row names the ported
// addFn/processFn, and its graphic drives the frame group AddMissile picks.
func TestMisdatRowsMatchWiredBodies(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)

	// when
	observed := make([]int, 0, len(wiredRows))
	for _, row := range wiredRows {
		observed = append(observed, frameGroupOf(t, rows, sprites, row))
	}

	// then
	for i, row := range wiredRows {
		t.Run(row.name, func(t *testing.T) {
			assertRegistered(t, row)
			assertRowColumns(t, rows, row)
			assertFrameGroupMatchesCorpus(t, sprites, row, observed[i])
			assertDrawFlagMatchesCorpus(t, rows, row)
		})
	}
}

// TestAddMissileInvisibleRowIsNotDrawn: a row flagged Invisible seeds
// DrawFlag false while a drawn row seeds true — the corpus rows ManaShield
// and Arrow pin both sides (tables/misdat.h:185-188).
func TestAddMissileInvisibleRowIsNotDrawn(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)

	// when
	invisible := addCorpusRow(t, rows, sprites, "ManaShield")
	drawn := addCorpusRow(t, rows, sprites, "Arrow")

	// then
	if invisible.DrawFlag {
		t.Fatal("ManaShield DrawFlag = true, want false (Invisible row)")
	}

	if !drawn.DrawFlag {
		t.Fatal("Arrow DrawFlag = false, want true")
	}
}

// TestAddMissileSeedsAnimationTiming: the row's frameDelay/frameLength land in
// AnimDelay/AnimLen at the frame group AddMissile picks (misdat.cpp:400-408).
// HolyBolt's delays differ between group 0 and the midir group, so the group
// actually indexes the columns.
func TestAddMissileSeedsAnimationTiming(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)
	row := misdatRow(t, rows, "HolyBolt")
	sprite := spriteRow(t, sprites, row.Graphic)
	ordinal, _ := spriteOrdinal(t, sprites, row.Graphic)
	group := int(missiles.West)

	list := &missiles.MissileList{}
	args := missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.West,
		Type:            missiles.HolyBolt,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		Flags:           row.Flags,
		AnimFrameAmount: sprite.NumFrames,
		FrameDelay:      sprite.FrameDelay,
		FrameLength:     sprite.FrameLength,
	}

	// when
	m := list.AddMissile(newSilentWorld(), args)

	// then
	if m == nil {
		t.Fatal("AddMissile returned nil for the HolyBolt row")
	}

	if sprite.FrameDelay[0] == sprite.FrameDelay[group] {
		t.Fatalf("corpus row keeps frameDelay %d across groups; pick another row",
			sprite.FrameDelay[0])
	}

	if m.MFNum != group {
		t.Fatalf("frame group = %d, want %d", m.MFNum, group)
	}

	if m.AnimDelay != sprite.FrameDelay[group] {
		t.Fatalf("AnimDelay = %d, want %d for group %d",
			m.AnimDelay, sprite.FrameDelay[group], group)
	}

	if m.AnimLen != sprite.FrameLength[group] {
		t.Fatalf("AnimLen = %d, want %d", m.AnimLen, sprite.FrameLength[group])
	}
}

// TestWiredRowsCastSoundPlumbsThroughAddMissile: every wired id fires the
// castSound column of its own corpus row at the launch tile
// (missiles.cpp:2840-2845), so the corpus text reaches the sink unchanged,
// and a row with an empty column (Arrow, MagmaBall, BloodStar) stays silent
// (addmissile.go:56).
func TestWiredRowsCastSoundPlumbsThroughAddMissile(t *testing.T) {
	// given
	rows := loadMisdat(t)

	// when
	played := make([]string, 0, len(wiredRows))
	for _, row := range wiredRows {
		played = append(played, castSoundOf(t, rows, loadSprites(t), row))
	}

	// then
	for i, row := range wiredRows {
		t.Run(row.name, func(t *testing.T) {
			corpus := misdatRow(t, rows, row.name)

			if played[i] != corpus.CastSound {
				t.Fatalf("cast sound = %q, want %q", played[i], corpus.CastSound)
			}
		})
	}
}

// TestWiredRowsHitSoundLandsOnTheMissile: AddMissile copies the row's
// hitSound onto the missile so the collision tail can play it at the impact
// tile (missiles.cpp:549-550); rows without one leave the field empty, which
// is the port's SfxID::None.
func TestWiredRowsHitSoundLandsOnTheMissile(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)

	// when
	seeded := make([]string, 0, len(wiredRows))
	for _, row := range wiredRows {
		seeded = append(seeded, seededHitSoundOf(t, rows, sprites, row))
	}

	// then
	for i, row := range wiredRows {
		t.Run(row.name, func(t *testing.T) {
			corpus := misdatRow(t, rows, row.name)

			if seeded[i] != corpus.HitSound {
				t.Fatalf("missile hitSound = %q, want the row's %q", seeded[i], corpus.HitSound)
			}
		})
	}
}

// TestRegistryInventory pins the slice-4 accounting: the registry covers one
// behavior per misdat row (Arrow..LastDiablo), exactly the wiredRows entries
// dispatch ported bodies, and everything else is still a no-op stub. A row
// ported without listing it here fails the wired-row checks above, and a row
// listed without porting it fails this count.
func TestRegistryInventory(t *testing.T) {
	// given
	rows := loadMisdat(t)
	wired := len(wiredRows)
	total := int(missiles.LastDiablo) + 1

	// when
	registered := missiles.RegistryCount()

	// then
	if registered != total {
		t.Fatalf("registry holds %d behaviors, want one per row (%d)", registered, total)
	}

	if len(rows) != total {
		t.Fatalf("corpus holds %d rows, want %d", len(rows), total)
	}

	if wired != 63 {
		t.Fatalf("wired rows = %d, want 63; update this count with the slice", wired)
	}

	if stubs := total - wired; stubs != 5 {
		t.Fatalf("stub rows = %d, want 5 (rows not yet ported)", stubs)
	}
}

// addCorpusRow appends one corpus row through AddMissile.
func addCorpusRow(t *testing.T, rows []misdat.MissileData,
	sprites []misdat.MissileSprite, name string,
) *missiles.Missile {
	t.Helper()

	row := misdatRow(t, rows, name)
	ordinal, _ := spriteOrdinal(t, sprites, row.Graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(castWorld(), missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   missiles.West,
		Type:    missiles.Phasing,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: missiles.GraphicID(ordinal),
		Flags:   row.Flags,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", name)
	}

	return m
}

// spriteRow finds a sprite row by id. An invisible row leaves the graphic
// column empty (misdat.tsv:41, Infravision) — that empty name is the corpus
// spelling of "no sprite" and yields the zero row.
func spriteRow(t *testing.T, sprites []misdat.MissileSprite, name string) misdat.MissileSprite {
	t.Helper()

	if name == "" {
		return misdat.MissileSprite{}
	}

	for _, sprite := range sprites {
		if sprite.ID == name {
			return sprite
		}
	}

	t.Fatalf("missile_sprites has no %s row", name)

	return misdat.MissileSprite{}
}

// loadMisdat parses the committed base misdat corpus.
func loadMisdat(t *testing.T) []misdat.MissileData {
	t.Helper()

	data, err := os.ReadFile(misdatPath)
	if err != nil {
		t.Fatalf("read %s: %v", misdatPath, err)
	}

	rows, err := misdat.ParseMisdat(data)
	if err != nil {
		t.Fatalf("misdat.ParseMisdat: %v", err)
	}

	return rows
}

// loadSprites parses the committed base missile_sprites corpus.
func loadSprites(t *testing.T) []misdat.MissileSprite {
	t.Helper()

	data, err := os.ReadFile(spritesPath)
	if err != nil {
		t.Fatalf("read %s: %v", spritesPath, err)
	}

	rows, err := misdat.ParseMissileSprites(data)
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites: %v", err)
	}

	return rows
}

// misdatRow finds a corpus row by id.
func misdatRow(t *testing.T, rows []misdat.MissileData, id string) misdat.MissileData {
	t.Helper()

	for _, row := range rows {
		if row.ID == id {
			return row
		}
	}

	t.Fatalf("misdat has no %s row", id)

	return misdat.MissileData{}
}

// spriteOrdinal returns the sprite's data row index — the MissileGraphicID
// ordinal — and its animFAmt (numFrames).
func spriteOrdinal(t *testing.T, sprites []misdat.MissileSprite,
	name string,
) (ordinal, frames int) {
	t.Helper()

	if name == "" {
		return 0, 0
	}

	for i, sprite := range sprites {
		if sprite.ID == name {
			return i, sprite.NumFrames
		}
	}

	t.Fatalf("missile_sprites has no %s row", name)

	return 0, 0
}

// frameGroupOf runs one row's graphic through AddMissile on an unwired id so
// the stub addFn leaves the frame-group pick in place, and checks it against
// the corpus animFAmt.
func frameGroupOf(t *testing.T, rows []misdat.MissileData,
	sprites []misdat.MissileSprite, row wiredRow,
) int {
	t.Helper()

	graphic := misdatRow(t, rows, row.name).Graphic
	ordinal, frames := spriteOrdinal(t, sprites, graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(castWorld(), missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.West,
		Type:            missiles.Healing,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		AnimFrameAmount: frames,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", row.name)
	}

	if frames >= 8 && m.MFNum != int(missiles.West) {
		t.Fatalf("%s: %d frame groups kept group %d, want %d",
			row.name, frames, m.MFNum, int(missiles.West))
	}

	if frames < 8 && m.MFNum != 0 {
		t.Fatalf("%s: %d frame groups followed midir to %d, want group 0",
			row.name, frames, m.MFNum)
	}

	return m.MFNum
}

// assertRowColumns checks the addFn/processFn/graphic of one corpus row.
func assertRowColumns(t *testing.T, rows []misdat.MissileData, row wiredRow) {
	t.Helper()

	got := misdatRow(t, rows, row.name)

	if got.AddFn != row.addFn {
		t.Fatalf("%s addFn = %s, want %s", row.name, got.AddFn, row.addFn)
	}

	if got.ProcessFn != row.processFn {
		t.Fatalf("%s processFn = %s, want %s", row.name, got.ProcessFn, row.processFn)
	}

	if got.Graphic != row.graphic {
		t.Fatalf("%s graphic = %s, want %s", row.name, got.Graphic, row.graphic)
	}
}

// assertFrameGroupMatchesCorpus re-derives the expectation from the corpus so
// the observed pick cannot drift from the committed data.
func assertFrameGroupMatchesCorpus(t *testing.T, sprites []misdat.MissileSprite,
	row wiredRow, observed int,
) {
	t.Helper()

	_, frames := spriteOrdinal(t, sprites, row.graphic)

	want := int(missiles.West)
	if frames < 8 {
		want = 0
	}

	if observed != want {
		t.Fatalf("%s frame group = %d, want %d for %d frame groups",
			row.name, observed, want, frames)
	}
}

// assertDrawFlagMatchesCorpus: AddMissile's seeded DrawFlag is the row's
// isDrawn() — false only for rows flagged Invisible (tables/misdat.h:185-188).
func assertDrawFlagMatchesCorpus(t *testing.T, rows []misdat.MissileData, row wiredRow) {
	t.Helper()

	misdatRow := misdatRow(t, rows, row.name)

	_, frames := spriteOrdinal(t, loadSprites(t), misdatRow.Graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(castWorld(), missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   missiles.West,
		Type:    row.id,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: missiles.GraphicID(graphicOf(t, misdatRow.Graphic)),
		Flags:   misdatRow.Flags,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", row.name)
	}

	want := true
	for _, flag := range misdatRow.Flags {
		if flag == "Invisible" {
			want = false
		}
	}

	if m.DrawFlag != want {
		t.Fatalf("%s DrawFlag = %v, want %v (flags %v, %d frame groups)",
			row.name, m.DrawFlag, want, misdatRow.Flags, frames)
	}
}

// graphicOf resolves a sprite name to its MissileGraphicID ordinal.
func graphicOf(t *testing.T, name string) int {
	t.Helper()

	ordinal, _ := spriteOrdinal(t, loadSprites(t), name)

	return ordinal
}

// seededHitSoundOf appends one corpus row and reports the hitSound the missile
// carries afterwards.
func seededHitSoundOf(t *testing.T, rows []misdat.MissileData,
	sprites []misdat.MissileSprite, row wiredRow,
) string {
	t.Helper()

	corpus := misdatRow(t, rows, row.name)
	sprite := spriteRow(t, sprites, corpus.Graphic)
	ordinal, _ := spriteOrdinal(t, sprites, corpus.Graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(castWorld(), missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.South,
		Type:            row.id,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		Flags:           corpus.Flags,
		CastSound:       corpus.CastSound,
		HitSound:        corpus.HitSound,
		AnimFrameAmount: sprite.NumFrames,
		FrameDelay:      sprite.FrameDelay,
		FrameLength:     sprite.FrameLength,
	})
	if m == nil {
		t.Fatalf("AddMissile returned nil for %s", row.name)
	}

	return m.HitSound
}

// castSoundOf appends one corpus row with its own castSound and reports what
// the sink heard.
func castSoundOf(t *testing.T, rows []misdat.MissileData, sprites []misdat.MissileSprite, row wiredRow) string {
	t.Helper()

	corpus := misdatRow(t, rows, row.name)
	sprite := spriteRow(t, sprites, corpus.Graphic)
	ordinal, _ := spriteOrdinal(t, sprites, corpus.Graphic)
	world, heard := newRecordingWorld()
	world.AddSource = castWorld().AddSource

	list := &missiles.MissileList{}
	list.AddMissile(world, missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.South,
		Type:            row.id,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		Flags:           corpus.Flags,
		CastSound:       corpus.CastSound,
		HitSound:        corpus.HitSound,
		AnimFrameAmount: sprite.NumFrames,
		FrameDelay:      sprite.FrameDelay,
		FrameLength:     sprite.FrameLength,
	})

	return *heard
}

// newRecordingWorld builds a silent world that records the last PlaySfxLoc name.
func newRecordingWorld() (*missiles.HitWorld, *string) {
	heard := new(string)
	world := newSilentWorld()
	world.ProcessCallbacks.PlaySfxLoc = func(name string, _, _ int) { *heard = name }

	return world, heard
}

// assertRegistered fails when a row is listed as wired but the dispatch map
// still hands out the no-op stub (behaviors only holds ported rows).
func assertRegistered(t *testing.T, row wiredRow) {
	t.Helper()

	if _, ok := missiles.BehaviorFor(row.id); !ok {
		t.Fatalf("%s is listed as wired but the registry still stubs it", row.name)
	}
}
