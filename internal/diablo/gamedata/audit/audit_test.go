//nolint:testpackage // internal-state test: pokes unexported identifiers
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/objdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/textdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/towners"
)

const (
	manifestFile = "audit_manifest.json"
	testdataDir  = "../../../testdata/gamedata"
)

var goldenPaths = map[string]string{
	"monstdat/base":        "monstdat/monstdat.tsv",
	"monstdat/hf":          "monstdat/hf-monstdat.tsv",
	"unique_monstdat/base": "monstdat/unique_monstdat.tsv",
	"misdat/base":          "misdat/misdat.tsv",
	"misdat/hf":            "misdat/hf-misdat.tsv",
	"effects/base":         "audio/effects.tsv",
	"effects/hf":           "audio/hf-effects.tsv",
	"textdat/base":         "textdat/textdat.tsv",
	"textdat/unused":       "textdat/textdat-unused.tsv",
	"towners/base":         "towners/towners.tsv",
	"towners/hf":           "towners/hf-towners.tsv",
	"quest_dialog/base":    "towners/quest_dialog.tsv",
	"quest_dialog/hf":      "towners/hf-quest_dialog.tsv",
	"objdat/base":          "objdat/objdat.tsv",
	"questdat/base":        "quests/questdat.tsv",
}

type rowCounter func(data []byte) (int, error)

var rowCounters = map[string]rowCounter{
	"monstdat": func(data []byte) (int, error) {
		rows, err := monstdat.ParseMonstdat(data)
		return len(rows), err
	},
	"unique_monstdat": func(data []byte) (int, error) {
		rows, err := monstdat.ParseUniqueMonstdat(data)
		return len(rows), err
	},
	"misdat": func(data []byte) (int, error) {
		rows, err := misdat.ParseMisdat(data)
		return len(rows), err
	},
	"effects": func(data []byte) (int, error) {
		rows, err := audio.ParseEffects(data)
		return len(rows), err
	},
	"textdat": func(data []byte) (int, error) {
		rows, err := textdat.ParseTextdat(data)
		return len(rows), err
	},
	"towners": func(data []byte) (int, error) {
		rows, err := towners.ParseTowners(data)
		return len(rows), err
	},
	"quest_dialog": func(data []byte) (int, error) {
		dialog, err := towners.ParseQuestDialog(data)
		return len(dialog.Rows), err
	},
	"objdat": func(data []byte) (int, error) {
		rows, err := objdat.ParseObjdat(data)
		return len(rows), err
	},
	"questdat": func(data []byte) (int, error) {
		rows, err := quests.ParseQuestdat(data)
		return len(rows), err
	},
}

func TestRowCountsMatchManifest(t *testing.T) {
	// given
	m, err := loadManifest(manifestFile)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(m.Entries) == 0 {
		t.Fatal("manifest has no entries")
	}

	// when
	err = checkRowCounts(t, m)

	// then
	if err != nil {
		t.Fatalf("row count mismatch: %v", err)
	}
}

func TestFingerprintsMatch(t *testing.T) {
	// given
	m, err := loadManifest(manifestFile)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for _, e := range m.Entries {
		t.Run(e.Table+"/"+e.Variant, func(t *testing.T) {
			// when
			rows := rawRows(t, tsvPath(e))

			// then
			if len(rows) == 0 {
				t.Fatal("golden has no rows")
			}
			if !reflect.DeepEqual(rows[0], e.FirstRow) {
				t.Errorf("first row mismatch for table %q variant %q:\n got %q\nwant %q",
					e.Table, e.Variant, rows[0], e.FirstRow)
			}
			last := rows[len(rows)-1]
			if !reflect.DeepEqual(last, e.LastRow) {
				t.Errorf("last row mismatch for table %q variant %q:\n got %q\nwant %q",
					e.Table, e.Variant, last, e.LastRow)
			}
		})
	}
}

func TestCorruptedManifestFailsWithTableAndVariant(t *testing.T) {
	// given
	m, err := loadManifest(manifestFile)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(m.Entries) == 0 {
		t.Fatal("manifest has no entries")
	}
	target := m.Entries[0]
	m.Entries[0].WantRows++

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal corrupted manifest: %v", err)
	}
	path := filepath.Join(t.TempDir(), "audit_manifest.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write corrupted manifest: %v", err)
	}
	corrupted, err := loadManifest(path)
	if err != nil {
		t.Fatalf("load corrupted manifest: %v", err)
	}

	// when
	err = checkRowCounts(t, corrupted)

	// then
	if err == nil {
		t.Fatal("expected the corrupted manifest to fail the row count check")
	}
	msg := err.Error()
	if !strings.Contains(msg, target.Table) || !strings.Contains(msg, target.Variant) {
		t.Fatalf("error %q does not name table %q variant %q", msg, target.Table, target.Variant)
	}
}

func tsvPath(e Entry) string {
	return goldenPaths[e.Table+"/"+e.Variant]
}

func readTSV(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testdataDir, rel))
	if err != nil {
		t.Fatalf("read %q: %v", rel, err)
	}
	return data
}

func rowsFor(t *testing.T, e Entry) int {
	t.Helper()
	counter, ok := rowCounters[e.Table]
	if !ok {
		t.Fatalf("unknown table %q", e.Table)
	}
	n, err := counter(readTSV(t, tsvPath(e)))
	if err != nil {
		t.Fatalf("parse %s %q: %v", e.Table, e.Variant, err)
	}
	return n
}

func rawRows(t *testing.T, rel string) [][]string {
	t.Helper()
	lines := strings.Split(string(readTSV(t, rel)), "\n")
	var rows [][]string
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

func checkRowCounts(t *testing.T, m Manifest) error {
	t.Helper()
	for _, e := range m.Entries {
		got := rowsFor(t, e)
		if got != e.WantRows {
			return fmt.Errorf("table %q variant %q: got %d rows, want %d", e.Table, e.Variant, got, e.WantRows)
		}
	}
	return nil
}
