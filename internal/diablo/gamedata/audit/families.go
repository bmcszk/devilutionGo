package audit

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/objdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/textdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/towners"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// Families returns the registry. Every .tsv under the corpus roots must be
// declared by exactly one glob; a new file that is not declared fails the
// audit by name (closed world), and a declaration with no file fails too.
func Families() []Family {
	families := make([]Family, 0, 64)
	families = append(families, gameFamilies()...)
	families = append(families, upstreamFamilies()...)

	return append(families, declaredNonTables()...)
}

// gameFamilies declares the tables the game loads, base rows and hf overrides.
func gameFamilies() []Family {
	return []Family{
		game("audio/effects.tsv", "effects", "base", effectKeys),
		game("audio/hf-effects.tsv", "effects", "hf", effectKeys),
		game("audio/effects-unused.tsv", "effects", "unused", effectKeys),
		game("audio/hf-effects-unused.tsv", "effects", "hf_unused", effectKeys),
		game("experience/Experience.tsv", "Experience", "base", tableKeys),
		game("items/itemdat.tsv", "itemdat", "base", itemDatKeys),
		game("items/item_prefixes.tsv", "item_prefixes", "base", prefixKeys),
		game("items/item_suffixes.tsv", "item_suffixes", "base", suffixKeys),
		game("items/unique_itemdat.tsv", "unique_itemdat", "base", uniqueItemKeys),
		game("items/hf-item_prefixes.tsv", "item_prefixes", "hf", prefixKeys),
		game("items/hf-item_suffixes.tsv", "item_suffixes", "hf", suffixKeys),
		game("items/hf-unique_itemdat.tsv", "unique_itemdat", "hf", uniqueItemKeys),
		game("misdat/misdat.tsv", "misdat", "base", misdatKeys),
		game("misdat/hf-misdat.tsv", "misdat", "hf", misdatKeys),
		game("misdat/missile_sprites.tsv", "missile_sprites", "base", spriteKeys),
		game("misdat/hf-missile_sprites.tsv", "missile_sprites", "hf", spriteKeys),
		game("monstdat/monstdat.tsv", "monstdat", "base", monstdatKeys),
		game("monstdat/hf-monstdat.tsv", "monstdat", "hf", monstdatKeys),
		game("monstdat/unique_monstdat.tsv", "unique_monstdat", "base", uniqueMonstdatKeys),
		game("objdat/objdat.tsv", "objdat", "base", objdatKeys),
		game("quests/questdat.tsv", "questdat", "base", questdatKeys),
		gameNamed("spells/spelldat.tsv", "spelldat", "base", 1, spellKeys),
		game("textdat/textdat.tsv", "textdat", "base", textdatKeys),
		game("textdat/textdat-unused.tsv", "textdat", "unused", textdatKeys),
		game("towners/towners.tsv", "towners", "base", townerKeys),
		game("towners/hf-towners.tsv", "towners", "hf", townerKeys),
		game("towners/quest_dialog.tsv", "quest_dialog", "base", questDialogKeys),
		game("towners/hf-quest_dialog.tsv", "quest_dialog", "hf", questDialogKeys),
	}
}

// upstreamFamilies declares the mirrored upstream corpus: the same tables
// before the game-side split, plus the per-class key/value files.
func upstreamFamilies() []Family {
	return []Family{
		upstream("upstream/classes/classdat.tsv", "classdat", "base", classKeys),
		upstream("upstream/classes/*/attributes.tsv", "attributes", "base", tableKeys),
		upstream("upstream/classes/*/animations.tsv", "animations", "base", tableKeys),
		upstream("upstream/classes/*/sounds.tsv", "sounds", "base", tableKeys),
		upstream("upstream/classes/*/sprites.tsv", "sprites", "base", tableKeys),
		upstream("upstream/classes/*/starting_loadout.tsv", "starting_loadout", "base", tableKeys),
		upstream("upstream/Experience.tsv", "Experience", "base", tableKeys),
		upstream("upstream/items/itemdat.tsv", "itemdat", "base", itemDatKeys),
		upstream("upstream/items/item_prefixes.tsv", "item_prefixes", "base", prefixKeys),
		upstream("upstream/items/item_suffixes.tsv", "item_suffixes", "base", suffixKeys),
		upstream("upstream/items/unique_itemdat.tsv", "unique_itemdat", "base", uniqueItemKeys),
		upstream("upstream/missiles/misdat.tsv", "misdat", "base", misdatKeys),
		upstream("upstream/missiles/missile_sprites.tsv", "missile_sprites", "base", spriteKeys),
		upstream("upstream/monsters/monstdat.tsv", "monstdat", "base", monstdatKeys),
		upstream("upstream/monsters/unique_monstdat.tsv", "unique_monstdat", "base", uniqueMonstdatKeys),
		upstream("upstream/objects/objdat.tsv", "objdat", "base", objdatKeys),
		upstream("upstream/quests/questdat.tsv", "questdat", "base", questdatKeys),
		upstream("upstream/sound/effects.tsv", "effects", "base", effectKeys),
		upstream("upstream/sound/effects-unused.tsv", "effects", "unused", effectKeys),
		upstreamKeyCol("upstream/spells/spelldat.tsv", "spelldat", "base", 1, spellKeys),
		upstream("upstream/text/textdat.tsv", "textdat", "base", textdatKeys),
		upstream("upstream/text/textdat-unused.tsv", "textdat", "unused", textdatKeys),
		upstream("upstream/towners/quest_dialog.tsv", "quest_dialog", "base", questDialogKeys),
		upstream("upstream/towners/towners.tsv", "towners", "base", townerKeys),
	}
}

// declaredNonTables keeps the world closed: goldens and loader probes are
// declared with the reason they carry no parser expectation.
func declaredNonTables() []Family {
	return []Family{
		{
			Corpus: CorpusGamedata, Glob: "lighttables/*.tsv", Table: "lighttables", Variant: "golden",
			Status: StatusGolden,
			Note:   "code-produced light golden (internal/light), read by tests not by a loader",
		},
		{
			Corpus: CorpusGamedata, Glob: "items/affix_pool_probe.tsv", Table: "affix_pool_probe",
			Variant: "probe", Status: StatusProbe,
			Note: "affix pool probe dump (itemgen tests), not a game table",
		},
		{
			Corpus: CorpusUpstream, Glob: "*.tsv", Table: "txtdata_probe", Variant: "probe",
			Status: StatusProbe,
			Note:   "loader probe fixture (BOM/CRLF/CR/empty/no trailing LF) or the fixture manifest",
		},
		{
			Corpus: CorpusUpstream, Glob: "golden/*.tsv", Table: "txtdata_probe", Variant: "synthetic",
			Status: StatusProbe,
			Note:   "synthetic schema golden for the txtdata parser, not a game table",
		},
	}
}

// game declares one game-corpus table whose identity is its first column.
func game(glob, table, variant string, parse Parser) Family {
	return Family{Corpus: CorpusGamedata, Glob: glob, Table: table, Variant: variant, Parse: parse}
}

// gameNamed declares one game-corpus table with an explicit identity column.
func gameNamed(glob, table, variant string, keyCol int, parse Parser) Family {
	family := game(glob, table, variant, parse)
	family.KeyCol = keyCol

	return family
}

// upstream declares one mirrored upstream table whose identity is its first column.
func upstream(glob, table, variant string, parse Parser) Family {
	return Family{Corpus: CorpusUpstream, Glob: glob, Table: table, Variant: variant, Parse: parse}
}

// upstreamKeyCol declares one mirrored upstream table with an explicit identity column.
func upstreamKeyCol(glob, table, variant string, keyCol int, parse Parser) Family {
	family := upstream(glob, table, variant, parse)
	family.KeyCol = keyCol

	return family
}

// parseKeys runs one loader and projects the identity out of every row.
func parseKeys[T any](data []byte, load func([]byte) ([]T, error), key func(T) string) ([]string, error) {
	rows, err := load(data)
	if err != nil {
		return nil, fmt.Errorf("parse table: %w", err)
	}

	keys := make([]string, len(rows))
	for i, row := range rows {
		keys[i] = key(row)
	}

	return keys, nil
}

// effectKeys reads effects.tsv through its loader (id|flags|path).
func effectKeys(data []byte) ([]string, error) {
	return parseKeys(data, audio.ParseEffects, func(row audio.Effect) string { return row.ID })
}

// itemDatKeys reads itemdat.tsv through its loader.
func itemDatKeys(data []byte) ([]string, error) {
	return parseKeys(data, itemdat.LoadItemDat, func(row itemdat.ItemData) string { return row.ID })
}

// prefixKeys reads item_prefixes.tsv through its loader.
func prefixKeys(data []byte) ([]string, error) {
	return parseKeys(data, itemdat.LoadItemPrefixes, func(row itemdat.ItemPrefix) string { return row.Name })
}

// suffixKeys reads item_suffixes.tsv through its loader.
func suffixKeys(data []byte) ([]string, error) {
	return parseKeys(data, itemdat.LoadItemSuffixes, func(row itemdat.ItemSuffix) string { return row.Name })
}

// uniqueItemKeys reads unique_itemdat.tsv through its loader.
func uniqueItemKeys(data []byte) ([]string, error) {
	return parseKeys(data, itemdat.LoadUniqueItemDat, func(row itemdat.UniqueItemData) string { return row.Name })
}

// misdatKeys reads misdat.tsv through its loader.
func misdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, misdat.ParseMisdat, func(row misdat.MissileData) string { return row.ID })
}

// spriteKeys reads missile_sprites.tsv through its loader.
func spriteKeys(data []byte) ([]string, error) {
	return parseKeys(data, misdat.ParseMissileSprites, func(row misdat.MissileSprite) string { return row.ID })
}

// monstdatKeys reads monstdat.tsv through its loader.
func monstdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, monstdat.ParseMonstdat, func(row monstdat.MonsterData) string { return row.ID })
}

// uniqueMonstdatKeys reads unique_monstdat.tsv through its loader.
func uniqueMonstdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, monstdat.ParseUniqueMonstdat, func(row []string) string { return cell(row, 0) })
}

// objdatKeys reads objdat.tsv through its loader.
func objdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, objdat.ParseObjdat, func(row objdat.ObjectData) string { return row.ID })
}

// questdatKeys reads questdat.tsv through its loader.
func questdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, quests.ParseQuestdat, func(row quests.QuestRow) string { return row.ID })
}

// textdatKeys reads textdat.tsv through its loader.
func textdatKeys(data []byte) ([]string, error) {
	return parseKeys(data, textdat.ParseTextdat, func(row textdat.TextRow) string { return row.TxtStrID })
}

// townerKeys reads towners.tsv through its loader.
func townerKeys(data []byte) ([]string, error) {
	return parseKeys(data, towners.ParseTowners, func(row towners.Towner) string { return row.Type })
}

// questDialogKeys reads quest_dialog.tsv through its loader.
func questDialogKeys(data []byte) ([]string, error) {
	dialog, err := towners.ParseQuestDialog(data)
	if err != nil {
		return nil, fmt.Errorf("parse quest_dialog: %w", err)
	}

	keys := make([]string, len(dialog.Rows))
	for i, row := range dialog.Rows {
		keys[i] = cell(row, 0)
	}

	return keys, nil
}

// spellKeys reads spelldat.tsv through its loader; the identity column is
// "name", because the loader validates "id" without storing it.
func spellKeys(data []byte) ([]string, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse spelldat table: %w", err)
	}

	rows, err := txtdata.LoadSpellData(tbl)
	if err != nil {
		return nil, fmt.Errorf("load spelldat: %w", err)
	}

	keys := make([]string, len(rows))
	for i, row := range rows {
		keys[i] = row.Name
	}

	return keys, nil
}

// classKeys reads classdat.tsv through its loader.
func classKeys(data []byte) ([]string, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse classdat table: %w", err)
	}

	rows, err := txtdata.LoadClasses(tbl)
	if err != nil {
		return nil, fmt.Errorf("load classdat: %w", err)
	}

	keys := make([]string, len(rows))
	for i, row := range rows {
		keys[i] = row.ClassName
	}

	return keys, nil
}

// tableKeys is the row identity of a plain txtdata table: its first column.
func tableKeys(data []byte) ([]string, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse txtdata table: %w", err)
	}

	keys := make([]string, len(tbl.Rows))
	for i, row := range tbl.Rows {
		keys[i] = cell(row, 0)
	}

	return keys, nil
}
