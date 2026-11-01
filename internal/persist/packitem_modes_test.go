package persist_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// uniqueBugDroppedFlag is CF_UPER1 (Source/items.h:163), the flag the
// unique-item morph bug leaves behind in _iCreateInfo and SetupAllItems drops
// (items.cpp:3555-3566).
const uniqueBugDroppedFlag uint16 = 256

// itemAvailabilityRows is the itemdat row count IsItemAvailable bounds against.
const itemAvailabilityRows = 200

func TestUnpackItemGoldenVectors(t *testing.T) {
	t.Parallel()

	for _, row := range loadItemPackGolden(t) {
		t.Run(row.suite+"/"+row.label, func(t *testing.T) {
			t.Parallel()

			// when the vector is expanded and packed back
			got := unpackThenPack(t, row.pack, envForSuite(row.suite))

			// then the record survives the round trip unchanged
			requireItemPack(t, got, row.pack)
		})
	}
}

func TestUnpackItemSpawnVectors(t *testing.T) {
	t.Parallel()

	for _, row := range loadItemPackModesGolden(t, "PackedSpawnItems") {
		t.Run(row.label, func(t *testing.T) {
			t.Parallel()

			// when a shareware save is expanded (pack_test.cpp:514)
			item, err := persist.UnpackItem(row.pack, envForSuite(row.suite))
			if err != nil {
				t.Fatalf("UnpackItem: %v", err)
			}

			// then the row is available in spawn and re-packs unchanged
			requireNotNone(t, item)

			requireItemPack(t, persist.PackItem(item, envForSuite(row.suite)), row.pack)
		})
	}
}

func TestUnpackItemDiabloMultiplayerVectors(t *testing.T) {
	t.Parallel()

	for _, row := range loadItemPackModesGolden(t, "PackedDiabloMPItems") {
		t.Run(row.label, func(t *testing.T) {
			t.Parallel()

			// when a multiplayer save is expanded (pack_test.cpp:559)
			item, err := persist.UnpackItem(row.pack, envForSuite(row.suite))
			if err != nil {
				t.Fatalf("UnpackItem: %v", err)
			}

			// then the row is available and re-packs unchanged
			requireNotNone(t, item)

			requireItemPack(t, persist.PackItem(item, envForSuite(row.suite)), row.pack)
		})
	}
}

func TestUnpackItemUniqueBug(t *testing.T) {
	t.Parallel()

	rows := loadItemPackModesGolden(t, "PackedUniqueBugItems")
	if len(rows) != 2 {
		t.Fatalf("got %d unique-bug vectors, want 2", len(rows))
	}

	bugged, fixed := rows[0].pack, rows[1].pack

	// given the morph-bug vector carries CF_UPER1 on top of CF_UPER15|CF_UNIQUE
	// in its create flags and the fixed vector carries only
	// CF_UPER15|CF_UNIQUE (pack_test.cpp:465-466)
	if bugged.CreateInfo&uniqueBugDroppedFlag == 0 {
		t.Fatalf("bugged CreateInfo = %#x, want CF_UPER1 set", bugged.CreateInfo)
	}

	if fixed.CreateInfo&uniqueBugDroppedFlag != 0 {
		t.Fatalf("fixed CreateInfo = %#x, want CF_UPER1 clear", fixed.CreateInfo)
	}

	// when the bugged vector is expanded and packed back
	item, err := persist.UnpackItem(bugged, persist.PackEnv{})
	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	got := persist.PackItem(item, persist.PackEnv{})

	// then the record layer is symmetric: both flags survive, because the
	// CF_UPER1 drop happens in SetupAllItems (items.cpp:3555-3566), not here.
	requireItemPack(t, got, bugged)

	// and the two vectors differ in exactly that one flag
	if got.CreateInfo^fixed.CreateInfo != uniqueBugDroppedFlag {
		t.Fatalf("CreateInfo delta = %#x, want %#x", got.CreateInfo^fixed.CreateInfo, uniqueBugDroppedFlag)
	}
}

func TestUnpackItemStripHellfireItems(t *testing.T) {
	t.Parallel()

	for _, row := range loadItemPackModesGolden(t, "PackedStripHellfireItems") {
		t.Run(row.label, func(t *testing.T) {
			t.Parallel()

			// when the Hellfire-exclusive record (pack_test.cpp:803) is
			// expanded by a Diablo build
			item, err := persist.UnpackItem(row.pack, envForSuite(row.suite))

			// then IsItemAvailable strips it (pack.cpp:305-308)
			requireHellfireStrip(t, item, err, row.pack)
		})
	}
}

func TestUnpackItemGoldCursors(t *testing.T) {
	t.Parallel()

	for _, row := range loadItemPackModesGolden(t, "PackedGoldItems") {
		t.Run(row.label, func(t *testing.T) {
			t.Parallel()

			// when a gold pile record is expanded
			item, err := persist.UnpackItem(row.pack, persist.PackEnv{})

			// then it carries the packed value and re-packs unchanged
			requireGoldPile(t, item, err, row)
		})
	}
}

// requireGoldPile checks one gold vector: the value survives and the record
// re-packs unchanged. The cursor breakpoints are 1000 and 2500
// (pack_test.cpp:862-878).
func requireGoldPile(t *testing.T, item entities.Item, err error, row itemPackRow) {
	t.Helper()

	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	want := map[string]int{"gold_small": 1000, "gold_medium": 1001, "gold_large": 2500}[row.label]
	if item.Ivalue != want {
		t.Fatalf("Ivalue = %d, want %d", item.Ivalue, want)
	}

	requireItemPack(t, persist.PackItem(item, persist.PackEnv{}), row.pack)
}

// requireHellfireStrip checks the two halves of the strip case: a Diablo read
// drops the record, a Hellfire read keeps it.
func requireHellfireStrip(t *testing.T, item entities.Item, err error, pack persist.ItemPack) {
	t.Helper()

	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	requireNone(t, item)

	hellfire, err := persist.UnpackItem(pack, envForSuite("PackedHellfireItems"))
	if err != nil {
		t.Fatalf("UnpackItem hellfire: %v", err)
	}

	requireNotNone(t, hellfire)
}

// envForSuite maps a pack_test.cpp suite name to the mode UnPackItem ran in.
//
// The itemdat index remaps (RemapItemIdxToDiablo/ToSpawn) are not ported yet,
// so the env leaves them nil: the vectors' own idx values are the
// reverse-remapped rows. See the residual note in the bean Proof of Work.
func envForSuite(suite string) persist.PackEnv {
	switch suite {
	case "PackedSpawnItems":
		return persist.PackEnv{Spawn: true, Available: availableIn(persist.PackEnv{Spawn: true})}
	case "PackedHellfireItems":
		return persist.PackEnv{Hellfire: true, Available: availableIn(persist.PackEnv{Hellfire: true})}
	case "PackedStripHellfireItems":
		return persist.PackEnv{Available: availableIn(persist.PackEnv{})}
	default:
		return persist.PackEnv{}
	}
}

// availableIn returns IsItemAvailable (items.cpp:2373-2404) bound to the mode.
// The row table is only used for its length upstream, so a zero slice of the
// itemdat row count is enough.
func availableIn(env persist.PackEnv) func(int16) bool {
	rows := make([]itemdat.ItemData, itemAvailabilityRows)

	drop := itemgen.DropEnv{Spawn: env.Spawn, Hellfire: env.Hellfire}

	return func(idx int16) bool {
		return itemgen.IsItemAvailable(rows, int(idx), drop)
	}
}

// unpackThenPack expands a record and packs it again.
func unpackThenPack(t *testing.T, pack persist.ItemPack, env persist.PackEnv) persist.ItemPack {
	t.Helper()

	item, err := persist.UnpackItem(pack, env)
	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	return persist.PackItem(item, env)
}

// requireItemPack fails unless the two records are field-identical.
func requireItemPack(t *testing.T, got, want persist.ItemPack) {
	t.Helper()

	if got != want {
		t.Fatalf("record = %+v, want %+v", got, want)
	}
}

// requireNotNone fails when the record was stripped to an empty slot
// (item.clear(), pack.cpp:307-309).
func requireNotNone(t *testing.T, item entities.Item) {
	t.Helper()

	if isCleared(item) {
		t.Fatal("item decoded as empty, want a real item")
	}
}

// requireNone fails when the record was not stripped.
func requireNone(t *testing.T, item entities.Item) {
	t.Helper()

	if !isCleared(item) {
		t.Fatalf("item = %+v, want an empty slot", item)
	}
}

// isCleared reports Item::isEmpty, which Item::clear sets by zeroing the type
// (inventory.EmptyItem).
func isCleared(item entities.Item) bool {
	return item.Itype == int8(inventory.ItemTypeNone)
}

// loadItemPackModesGolden reads the mode vectors of one suite from
// testdata/item_pack_modes_golden.tsv.
func loadItemPackModesGolden(t *testing.T, suite string) []itemPackRow {
	t.Helper()

	file, err := os.Open("testdata/item_pack_modes_golden.tsv")
	if err != nil {
		t.Fatalf("open mode golden: %v", err)
	}

	defer file.Close() //nolint:errcheck // read-only fixture

	rows := scanModeGolden(t, file, suite)
	if len(rows) == 0 {
		t.Fatalf("mode golden has no %s rows", suite)
	}

	return rows
}

// scanModeGolden parses every non-comment row of one suite.
func scanModeGolden(t *testing.T, file *os.File, suite string) []itemPackRow {
	t.Helper()

	var rows []itemPackRow

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" || !strings.HasPrefix(line, suite+"\t") {
			continue
		}

		rows = append(rows, parseItemPackRow(t, line))
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan mode golden: %v", err)
	}

	return rows
}
