package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJournalMissingFileIsFreshRun(t *testing.T) {
	// given an assets dir with no checkpoint yet
	out := t.TempDir()

	// when the journal is loaded
	loaded, err := loadJournal(out)
	// then it is empty and bound to that dir
	if err != nil {
		t.Fatalf("loadJournal: %v", err)
	}

	if loaded.Version != journalVersion || len(loaded.Archives) != 0 {
		t.Errorf("journal = %+v, want an empty version %d journal", loaded, journalVersion)
	}

	if loaded.Out != out {
		t.Errorf("journal out = %q, want %q", loaded.Out, out)
	}
}

func TestLoadJournalRejectsForeignVersion(t *testing.T) {
	// given a checkpoint written by another tool version
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(out, journalName), []byte(`{"version":99}`), 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	// when it is loaded
	_, err := loadJournal(out)

	// then the run stops instead of guessing what the file means
	if err == nil {
		t.Fatal("loadJournal accepted a foreign version")
	}
}

func TestJournalRoundTrip(t *testing.T) {
	// given a journal with one completed archive
	out := t.TempDir()
	state := emptyJournal(out)
	state.Archives["spawn.mpq"] = archiveState{Bytes: 4, Modified: 7, Files: []fileEntry{{Path: "spawn/a.clx"}}}

	// when it is saved and loaded again
	if err := saveJournal(out, state); err != nil {
		t.Fatalf("saveJournal: %v", err)
	}

	loaded, err := loadJournal(out)
	if err != nil {
		t.Fatalf("loadJournal: %v", err)
	}

	// then the recorded archive survives with its fingerprint
	if got := loaded.Archives["spawn.mpq"]; got.Bytes != 4 || got.Modified != 7 || len(got.Files) != 1 {
		t.Errorf("archive state = %+v, want the saved fingerprint", got)
	}
}

func TestResumeReusesCompletedArchive(t *testing.T) {
	// given a journal recording an archive and a file that is still intact
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\health.clx`: clxFixture(t)})
	out := t.TempDir()

	written, err := extractInto(t, archive, out)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	state := emptyJournal(out)
	state.Archives[archive] = written.state

	// when the archive is resumed
	recorded, resumed := state.resume(archive)

	// then the recorded files are reused
	if !resumed {
		t.Fatal("resume refused an untouched archive")
	}

	if len(recorded.Files) != 1 {
		t.Errorf("resumed files = %d, want 1", len(recorded.Files))
	}
}

func TestResumeRejectsModifiedOutput(t *testing.T) {
	// given a journal whose recorded output file was changed on disk
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\health.clx`: clxFixture(t)})
	out := t.TempDir()

	written, err := extractInto(t, archive, out)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	state := emptyJournal(out)
	state.Archives[archive] = written.state

	tampered := filepath.Join(out, filepath.FromSlash(written.state.Files[0].Path))
	if err := os.WriteFile(tampered, []byte("not the record"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	// when the archive is resumed
	_, resumed := state.resume(archive)

	// then it is extracted again instead of trusting the journal
	if resumed {
		t.Fatal("resume trusted a journal whose output no longer matches")
	}
}

func TestResumeRejectsMissingOutput(t *testing.T) {
	// given a journal whose recorded output file was deleted
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\health.clx`: clxFixture(t)})
	out := t.TempDir()

	written, err := extractInto(t, archive, out)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	state := emptyJournal(out)
	state.Archives[archive] = written.state

	if err := os.Remove(filepath.Join(out, filepath.FromSlash(written.state.Files[0].Path))); err != nil {
		t.Fatalf("remove output: %v", err)
	}

	// when the archive is resumed
	_, resumed := state.resume(archive)

	// then the missing file makes the archive stale
	if resumed {
		t.Fatal("resume accepted a journal with a missing output file")
	}
}

func TestResumeRejectsChangedArchive(t *testing.T) {
	// given a journal recorded for one archive's fingerprint
	state, entry := journalFixture(t, t.TempDir())
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\health.clx`: clxFixture(t)})

	info, err := os.Stat(archive)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	state.Archives[archive] = archiveState{
		Bytes:    info.Size(),
		Modified: info.ModTime().UnixNano() + 1,
		Files:    []fileEntry{entry},
	}

	// when the archive's recorded fingerprint does not match the file on disk
	_, resumed := state.resume(archive)

	// then the archive is extracted again
	if resumed {
		t.Fatal("resume accepted a journal recorded for different archive bytes")
	}
}

func TestResumeRejectsUnknownArchive(t *testing.T) {
	// given a journal with no record of this archive
	state := emptyJournal(t.TempDir())

	// when an archive the journal never saw is resumed
	_, resumed := state.resume("nowhere.mpq")

	// then nothing is reused
	if resumed {
		t.Fatal("resume invented a record for an unknown archive")
	}
}

// extractInto extracts archive into out, as the run does.
func extractInto(t *testing.T, archive, out string) (extraction, error) {
	t.Helper()

	extracted, err := extractArchive(archive, nil)
	if err != nil {
		return extraction{}, err
	}

	return extracted, writeEntries(out, extracted.state.Files)
}

// journalFixture writes one output file and returns a journal that recorded it.
func journalFixture(t *testing.T, out string) (*journal, fileEntry) {
	t.Helper()

	entry := fileEntry{Path: "spawn/data/health.clx", Bytes: 4, SHA256: digest([]byte("clx!"))}
	if err := writeOutput(filepath.Join(out, filepath.FromSlash(entry.Path)), []byte("clx!")); err != nil {
		t.Fatalf("write output: %v", err)
	}

	state := emptyJournal(out)

	return state, entry
}
