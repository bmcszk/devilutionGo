// Package persist_test: parity tests over the real upstream reference save of
// timerun demo 0 (committed as an xxd text fixture, pin e1ad0c2b6).
package persist_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

const (
	// referenceSaveHexPath is the committed xxd dump of
	// test/fixtures/timedemo/WarriorLevel1to2/demo_0_reference_spawn_0.sv.
	referenceSaveHexPath = "testdata/demo_0_reference_spawn_0.sv.hex"
	// referenceEntriesPath is the committed entry/size table of that archive.
	referenceEntriesPath = "testdata/demo_0_reference_spawn_0.entries.tsv"
	// referenceSaveBytes is the binary size of the reference archive at the
	// pin (sha256 78c0708a...).
	referenceSaveBytes = 574799
)

// referenceEntry is one row of the committed archive-entry table.
type referenceEntry struct {
	name    string
	encoded int
	decoded int
}

func TestReferenceArchiveLoads(t *testing.T) {
	t.Parallel()

	// given the committed reference archive of timerun demo 0
	store := referenceStore(t)

	// when its hero entry is read
	heroEntry, err := store.ReadEntry(persist.HeroEntryName)
	if err != nil {
		t.Fatalf("ReadEntry(hero): %v", err)
	}

	// then the entry is the codec image of one PlayerPack, not the plaintext
	// record (codec.cpp:132-138, pfile.cpp:177)
	if got, want := len(heroEntry), persist.HeroEntryLen(persist.PlayerPackSize); got != want {
		t.Fatalf("hero entry = %d bytes, want %d (codec_get_encoded_len(%d))",
			got, want, persist.PlayerPackSize)
	}

	// ...and it decodes to exactly one PlayerPack (pfile.cpp:628)
	hero, _, err := store.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got := len(hero.Pack); got != persist.PlayerPackSize {
		t.Fatalf("hero pack = %d bytes, want %d", got, persist.PlayerPackSize)
	}
}

func TestReferenceArchiveEntriesMatchCommittedSizes(t *testing.T) {
	t.Parallel()

	// given the committed entry table of the reference archive
	entries := readReferenceEntries(t)
	if len(entries) == 0 {
		t.Fatal("committed entry table is empty")
	}

	// when every row's entry is read back and codec-decoded
	sizes := readEntrySizes(t, referenceStore(t), entries)

	// then the encoded and decoded lengths match the committed row
	for i, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()

			assertEntrySizes(t, entry, sizes[i])
		})
	}
}

func TestCompareSavesReferenceAgainstItself(t *testing.T) {
	t.Parallel()

	// given the reference archive on both sides
	store := referenceStore(t)

	// when it is compared with itself
	result, err := persist.CompareSaves(store, store, demoOptions(false))
	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	// then every compared entry matches (pfile.cpp:495-503)
	if result.Status != persist.DemoSame {
		t.Fatalf("status = %v, want Same: %s", result.Status, result.Message)
	}

	if result.Message != "" {
		t.Fatalf("message = %q, want empty", result.Message)
	}
}

func TestCompareSavesReferenceAgainstPlaintextArchive(t *testing.T) {
	t.Parallel()

	// given the reference archive and an archive holding the SAME game data in
	// this store's export layout (encoded hero, plaintext game stream)
	reference := referenceStore(t)
	actual := plaintextTwinArchive(t, reference)

	// when they are compared
	result, err := persist.CompareSaves(actual, reference, demoOptions(false))
	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	// then they are Same: the comparator codec-decodes BOTH sides like
	// ReadArchive (pfile.cpp:619-630) and tolerates our plaintext export
	if result.Status != persist.DemoSame {
		t.Fatalf("status = %v, want Same: %s", result.Status, result.Message)
	}
}

func TestCompareSavesReferenceOneByteFlipNamesEntry(t *testing.T) {
	t.Parallel()

	// given the reference archive and its compared entries (pfile.cpp:473-484)
	entries := comparedEntries(readReferenceEntries(t))

	// when a copy with one flipped payload byte is compared, per entry
	results := flippedComparisons(t, entries)

	// then each report is a content difference naming that entry
	// (pfile.cpp:505-508)
	for i, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()

			assertDifferenceNames(t, entry.name, results[i])
		})
	}
}

// assertEntrySizes checks one read entry against its committed row.
func assertEntrySizes(t *testing.T, want, got referenceEntry) {
	t.Helper()

	if got.encoded != want.encoded {
		t.Fatalf("encoded = %d bytes, want %d", got.encoded, want.encoded)
	}

	if got.decoded != want.decoded {
		t.Fatalf("decoded = %d bytes, want %d", got.decoded, want.decoded)
	}
}

// assertDifferenceNames checks the content-difference wording of
// pfile.cpp:505-508 names one entry.
func assertDifferenceNames(t *testing.T, entry string, result persist.HeroCompareResult) {
	t.Helper()

	if result.Status != persist.DemoDifference {
		t.Fatalf("status = %v, want Difference", result.Status)
	}

	want := `file "` + entry + `" has different content.`
	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}

// readEntrySizes reads and codec-decodes every committed row's entry
// (pfile.cpp:619-630).
func readEntrySizes(t *testing.T, store persist.LegacySVStore, entries []referenceEntry) []referenceEntry {
	t.Helper()

	password := persist.HeroPassword(persist.SaveNamespace{Spawn: true})
	sizes := make([]referenceEntry, len(entries))

	for i, entry := range entries {
		encoded, err := store.ReadEntry(entry.name)
		if err != nil {
			t.Fatalf("ReadEntry(%q): %v", entry.name, err)
		}

		decoded, err := persist.DecodeHeroEntry(append([]byte(nil), encoded...), password)
		if err != nil {
			t.Fatalf("decode %q: %v", entry.name, err)
		}

		sizes[i] = referenceEntry{name: entry.name, encoded: len(encoded), decoded: len(decoded)}
	}

	return sizes
}

// flippedComparisons compares the reference against one copy per entry, each
// with that entry's decoded payload one byte changed.
func flippedComparisons(t *testing.T, entries []referenceEntry) []persist.HeroCompareResult {
	t.Helper()

	reference := referenceStore(t)
	results := make([]persist.HeroCompareResult, len(entries))

	for i, entry := range entries {
		actual := flippedArchive(t, reference, entry.name)

		result, err := persist.CompareSaves(actual, reference, demoOptions(false))
		if err != nil {
			t.Fatalf("CompareSaves(%q): %v", entry.name, err)
		}

		results[i] = result
	}

	return results
}

// spawnSingle is the mode the reference save was recorded in: shareware,
// single player, hence PASSWORD_SPAWN_SINGLE (pfile.cpp:638-643).
func spawnSingle() persist.GameMode { return persist.GameMode{Spawn: true} }

// referenceStore writes the committed reference archive into a temp dir and
// returns a store over it.
func referenceStore(t *testing.T) persist.LegacySVStore {
	t.Helper()

	raw := readReferenceSVBytes(t)

	path := filepath.Join(t.TempDir(), "demo_0_reference_single_0.sv")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write reference archive: %v", err)
	}

	return persist.LegacySVStore{Path: path, Mode: spawnSingle()}
}

// readReferenceSVBytes decodes the committed xxd dump of the reference save.
func readReferenceSVBytes(t *testing.T) []byte {
	t.Helper()

	dump, err := os.ReadFile(referenceSaveHexPath)
	if err != nil {
		t.Fatalf("read committed reference dump: %v", err)
	}

	raw, err := testsupport.DecodeXXD(dump)
	if err != nil {
		t.Fatalf("decode committed reference dump: %v", err)
	}

	if len(raw) != referenceSaveBytes {
		t.Fatalf("reference archive = %d bytes, want the pinned %d", len(raw), referenceSaveBytes)
	}

	return raw
}

// readReferenceEntries parses the committed entry table; the row count comes
// from the file, never from a literal.
func readReferenceEntries(t *testing.T) []referenceEntry {
	t.Helper()

	raw, err := os.ReadFile(referenceEntriesPath)
	if err != nil {
		t.Fatalf("read committed entry table: %v", err)
	}

	var entries []referenceEntry

	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entries = append(entries, parseReferenceEntry(t, i+1, line))
	}

	return entries
}

// parseReferenceEntry parses one "entry<TAB>encoded<TAB>decoded" row.
func parseReferenceEntry(t *testing.T, lineNo int, line string) referenceEntry {
	t.Helper()

	fields := strings.Split(line, "\t")
	if len(fields) != 3 {
		t.Fatalf("entry table line %d = %q, want 3 tab-separated fields", lineNo, line)
	}

	return referenceEntry{
		name:    fields[0],
		encoded: parseSize(t, lineNo, fields[1]),
		decoded: parseSize(t, lineNo, fields[2]),
	}
}

// parseSize parses one byte-count column.
func parseSize(t *testing.T, lineNo int, field string) int {
	t.Helper()

	size, err := strconv.Atoi(field)
	if err != nil {
		t.Fatalf("entry table line %d size %q: %v", lineNo, field, err)
	}

	return size
}

// decodedEntries codec-decodes every entry of a reference-shaped archive
// (pfile.cpp:619-630).
func decodedEntries(t *testing.T, store persist.LegacySVStore) map[string][]byte {
	t.Helper()

	entries := readReferenceEntries(t)
	password := persist.HeroPassword(persist.SaveNamespace{Spawn: true})
	decoded := make(map[string][]byte, len(entries))

	for _, entry := range entries {
		encoded, err := store.ReadEntry(entry.name)
		if err != nil {
			t.Fatalf("ReadEntry(%q): %v", entry.name, err)
		}

		payload, err := persist.DecodeHeroEntry(append([]byte(nil), encoded...), password)
		if err != nil {
			t.Fatalf("decode %q: %v", entry.name, err)
		}

		decoded[entry.name] = payload
	}

	return decoded
}

// plaintextTwinArchive rewrites the reference archive the way this store's
// export writes it today: codec-encoded hero, plaintext auxiliary streams.
func plaintextTwinArchive(t *testing.T, reference persist.LegacySVStore) persist.LegacySVStore {
	t.Helper()

	decoded := decodedEntries(t, reference)
	hero, ok := decoded[persist.HeroEntryName]
	if !ok {
		t.Fatalf("reference archive has no %q entry", persist.HeroEntryName)
	}

	return writeArchive(t, persist.HeroSave{
		Hero:  persist.HeroRecord{Pack: hero},
		Game:  decoded[persist.GameEntryName],
		Extra: auxiliary(decoded, plaintextEntry),
	})
}

// flippedArchive rewrites the reference archive with one decoded payload byte
// flipped, then re-encodes every entry as upstream does (loadsave.cpp:250-254).
func flippedArchive(t *testing.T, reference persist.LegacySVStore, entry string) persist.LegacySVStore {
	t.Helper()

	decoded := decodedEntries(t, reference)
	payload, ok := decoded[entry]
	if !ok {
		t.Fatalf("reference archive has no %q entry", entry)
	}

	payload[len(payload)-1] ^= 0xFF

	password := persist.HeroPassword(persist.SaveNamespace{Spawn: true})

	return writeArchive(t, persist.HeroSave{
		Hero: persist.HeroRecord{Pack: decoded[persist.HeroEntryName]},
		Game: encodedEntry(t, decoded[persist.GameEntryName], password),
		Extra: auxiliary(decoded, func(_ string, payload []byte) []byte {
			return encodedEntry(t, payload, password)
		}),
	})
}

// writeArchive writes a hero save into a fresh archive under a temp dir.
func writeArchive(t *testing.T, save persist.HeroSave) persist.LegacySVStore {
	t.Helper()

	path := filepath.Join(t.TempDir(), "demo_0_actual_single_0.sv")

	store := persist.LegacySVStore{Path: path, Mode: spawnSingle()}
	if err := store.WriteSave(save); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	return store
}

// auxiliary maps every entry except hero and game through encode.
func auxiliary(decoded map[string][]byte, encode func(name string, payload []byte) []byte,
) map[string][]byte {
	streams := make(map[string][]byte, len(decoded))

	for name, payload := range decoded {
		if name == persist.HeroEntryName || name == persist.GameEntryName {
			continue
		}

		streams[name] = encode(name, payload)
	}

	return streams
}

// plaintextEntry carries an entry through unchanged — the layout this store
// writes today (the write-side codec gap).
func plaintextEntry(_ string, payload []byte) []byte { return payload }

// encodedEntry codec-encodes one archive entry (codec.cpp:140-173).
func encodedEntry(t *testing.T, payload []byte, password string) []byte {
	t.Helper()

	encoded, err := persist.EncodeHeroEntry(payload, password)
	if err != nil {
		t.Fatalf("encode entry: %v", err)
	}

	return encoded
}

// comparedEntries narrows the committed entry table to the entries CompareSaves
// walks (pfile.cpp:473-484): hotkeys and heroitems live in the archive but are
// not comparison targets.
func comparedEntries(entries []referenceEntry) []referenceEntry {
	compared := make([]referenceEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.name == "hotkeys" || entry.name == "heroitems" {
			continue
		}

		compared = append(compared, entry)
	}

	return compared
}
