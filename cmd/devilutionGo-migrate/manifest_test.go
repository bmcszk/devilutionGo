package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestSummarizeIgnoresEntryOrder(t *testing.T) {
	// given the same two files in opposite order
	first := fileEntry{Path: "spawn/a.clx", SHA256: "aa", Bytes: 10}
	second := fileEntry{Path: "spawn/b.pal", SHA256: "bb", Bytes: 20}

	// when both orders are summarized
	forward, backward := summarize([]fileEntry{first, second}), summarize([]fileEntry{second, first})

	// then the hash describes the file set, not the walk order
	if forward.SHA256 != backward.SHA256 {
		t.Errorf("summary depends on order: %s vs %s", forward.SHA256, backward.SHA256)
	}

	if forward.Files != 2 || forward.Bytes != 30 {
		t.Errorf("summary = %+v, want 2 files and 30 bytes", forward)
	}
}

func TestSummarizeChangesWithContent(t *testing.T) {
	// when two versions of one file, with different content hashes, are summarized
	before := summarize([]fileEntry{{Path: "spawn/a.clx", SHA256: "aa"}})
	after := summarize([]fileEntry{{Path: "spawn/a.clx", SHA256: "ab"}})

	// then the summary differs, so a rerun that changes bytes cannot look
	// idempotent
	if before.SHA256 == after.SHA256 {
		t.Error("summary ignored a changed file hash")
	}
}

func TestTableCSVKeepsValuesVerbatim(t *testing.T) {
	// given a table with a quoted value and an empty cell
	table, err := assets.Load(bytes.NewReader([]byte("id\tname\tnote\n1\t\"quoted\"\t\n")))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// when it is rendered as CSV
	content, err := tableCSV(table)
	if err != nil {
		t.Fatalf("tableCSV: %v", err)
	}

	// then the CSV quotes where RFC 4180 requires it and keeps the empty cell
	if want := "id,name,note\n1,\"\"\"quoted\"\"\",\n"; string(content) != want {
		t.Errorf("csv = %q, want %q", content, want)
	}
}

func TestWriteManifestIsStable(t *testing.T) {
	// given one asset list
	listed := manifest{
		Tool:     manifestTool,
		Version:  manifestVersion,
		Archives: []archiveRef{{Path: "spawn.mpq", Bytes: 1024, Modified: 7, Files: 1}},
		Files: []fileEntry{
			{
				Path:    "spawn/a.clx",
				Archive: "spawn.mpq",
				Source:  `a\a.clx`,
				Family:  familySprite,
				Format:  "clx",
				Bytes:   4,
				SHA256:  "aa",
			},
		},
		Summary: summarize([]fileEntry{{Path: "spawn/a.clx", SHA256: "aa", Bytes: 4}}),
	}

	first := filepath.Join(t.TempDir(), "one")
	second := filepath.Join(t.TempDir(), "two")

	// when it is written twice into different dirs
	if err := writeManifest(first, listed); err != nil {
		t.Fatalf("write first: %v", err)
	}

	if err := writeManifest(second, listed); err != nil {
		t.Fatalf("write second: %v", err)
	}

	// then the bytes are identical and the file parses back
	firstBytes, err := os.ReadFile(filepath.Join(first, manifestName))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	secondBytes, err := os.ReadFile(filepath.Join(second, manifestName))
	if err != nil {
		t.Fatalf("read second: %v", err)
	}

	if !bytes.Equal(firstBytes, secondBytes) {
		t.Error("manifest bytes differ between two writes of the same list")
	}

	if parsed := readManifest(t, first); parsed.Summary.SHA256 == "" {
		t.Error("manifest round trip lost the summary")
	}
}

func TestBuildManifestRecordsArchiveFingerprint(t *testing.T) {
	// given a journal holding one completed archive
	state := emptyJournal(t.TempDir())
	state.Archives["spawn.mpq"] = archiveState{Bytes: 512, Modified: 99, Files: []fileEntry{{Path: "spawn/a.clx"}}}

	// when the manifest is built
	listed := buildManifest([]string{"spawn.mpq"}, state, []fileEntry{})

	// then the archive entry carries its fingerprint, so a reader can tell which
	// archive bytes the list describes
	if len(listed.Archives) != 1 {
		t.Fatalf("archives = %v", listed.Archives)
	}

	ref := listed.Archives[0]
	if ref.Bytes != 512 || ref.Modified != 99 || ref.Files != 1 {
		t.Errorf("archive ref = %+v, want the recorded fingerprint", ref)
	}
}
