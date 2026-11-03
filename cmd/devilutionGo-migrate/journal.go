package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// journalName is the checkpoint file inside the assets dir, so the resume
	// state travels with the tree it describes.
	journalName = ".migrate-state.json"
	// journalVersion guards against reading a journal an older tool wrote.
	journalVersion = 1
)

// archiveState is what one archive's extraction completed: the archive's
// fingerprint and every file it produced.
type archiveState struct {
	Bytes    int64       `json:"bytes"`
	Modified int64       `json:"modified_unix_nano"`
	Files    []fileEntry `json:"files"`
}

// journal is the checkpoint file. It is keyed by archive path; the fingerprint
// says whether the recorded files still describe that archive, so an archive
// that changed on disk is extracted again instead of silently reused.
type journal struct {
	Version  int                     `json:"version"`
	Archives map[string]archiveState `json:"archives"`
	Out      string                  `json:"out"`
}

// emptyJournal returns a journal for a fresh assets dir.
func emptyJournal(out string) *journal {
	return &journal{Version: journalVersion, Archives: map[string]archiveState{}, Out: out}
}

// loadJournal reads the checkpoint of an existing assets dir. A missing file is
// a first run, not an error; an unreadable or foreign journal is an error,
// because guessing would write a tree the journal no longer describes.
func loadJournal(out string) (*journal, error) {
	content, err := os.ReadFile(filepath.Join(out, journalName))
	if errors.Is(err, os.ErrNotExist) {
		return emptyJournal(out), nil
	}

	if err != nil {
		return nil, fmt.Errorf("read journal: %w", err)
	}

	loaded := &journal{}

	if err := json.Unmarshal(content, loaded); err != nil {
		return nil, fmt.Errorf("parse journal: %w", err)
	}

	if loaded.Version != journalVersion {
		return nil, fmt.Errorf("journal version %d, want %d: delete %s to start over",
			loaded.Version, journalVersion, filepath.Join(out, journalName))
	}

	if loaded.Archives == nil {
		loaded.Archives = map[string]archiveState{}
	}

	loaded.Out = out

	return loaded, nil
}

// saveJournal checkpoints the run; the journal is written last so a crash leaves
// the files it describes either incomplete on disk or absent from the journal,
// and the next run redoes that archive.
func saveJournal(out string, state *journal) error {
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode journal: %w", err)
	}

	if err := writeOutput(filepath.Join(out, journalName), append(content, '\n')); err != nil {
		return fmt.Errorf("write journal: %w", err)
	}

	return nil
}

// resume returns the recorded state for archive when it still describes the file
// on disk and every file it lists is intact; otherwise it reports false and the
// archive is extracted again.
func (j *journal) resume(archive string) (archiveState, bool) {
	recorded, known := j.Archives[archive]
	if !known {
		return archiveState{}, false
	}

	current, err := os.Stat(archive)
	if err != nil {
		return archiveState{}, false
	}

	if current.Size() != recorded.Bytes || current.ModTime().UnixNano() != recorded.Modified {
		return archiveState{}, false
	}

	for _, entry := range recorded.Files {
		if !j.intact(entry) {
			return archiveState{}, false
		}
	}

	return recorded, true
}

// intact reports whether a recorded output file is still in the assets dir with
// the hash the journal recorded for it.
func (j *journal) intact(entry fileEntry) bool {
	content, err := os.ReadFile(filepath.Join(j.Out, filepath.FromSlash(entry.Path)))
	if err != nil {
		return false
	}

	return digest(content) == entry.SHA256
}
