package persist

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// SaveNamespace selects the save-file family pfile uses
// (Source/pfile.cpp:78-97).
type SaveNamespace struct {
	// Spawn is gbIsSpawn (shareware).
	Spawn bool
	// Multiplayer is gbIsMultiplayer.
	Multiplayer bool
	// Extension is the effective save extension token, "sv" unless a mod
	// overrides it (pfile.cpp:74-77 GetSaveExtension).
	Extension string
}

// HeroSlots is the character-selection list: hero_names[MAX_CHARACTERS]
// (Source/pfile.cpp:69) plus the seen/slot bookkeeping the UI needs.
type HeroSlots struct {
	names [MaxCharacters]string
	used  [MaxCharacters]bool
}

// HeroSlotEntry is one occupied character-list entry: the save number and the
// name the selection dialog shows (_uiheroinfo, Source/DiabloUI/diabloui.h:60-72,
// filled per slot by pfile_ui_set_hero_infos, Source/pfile.cpp:695-719).
// Level, class, rank, attributes and hassaved come from the decoded pack
// (Game2UiPlayer, Source/pfile.cpp:211-222) and are not held at this layer.
type HeroSlotEntry struct {
	// Slot is saveNumber (pfile.cpp:706).
	Slot int
	// Name is hero_names[slot] (pfile.cpp:707).
	Name string
}

// PrimaryPrefix is the "single_"/"multi_"/"spawn_"/"share_" token of
// GetSavePath (pfile.cpp:81-88).
func (n SaveNamespace) PrimaryPrefix() string {
	switch {
	case n.Spawn && n.Multiplayer:
		return "share_"
	case n.Spawn:
		return "spawn_"
	case n.Multiplayer:
		return "multi_"
	default:
		return "single_"
	}
}

// HeroPath is GetSavePath(saveNum) (Source/pfile.cpp:79-92).
func (n SaveNamespace) HeroPath(dir string, slot int) (string, error) {
	return n.SavePath(dir, slot, "")
}

// SavePath is GetSavePath(saveNum, savePrefix) (pfile.cpp:79-98).
func (n SaveNamespace) SavePath(dir string, slot int, prefix string) (string, error) {
	if slot < 0 || slot >= MaxCharacters {
		return "", fmt.Errorf("persist: save slot %d out of range 0..%d: %w", slot, MaxCharacters-1,
			ErrSlotOutOfRange)
	}

	name := prefix + n.PrimaryPrefix() + strconv.Itoa(slot) + "." + n.extension()

	return filepath.Join(dir, name), nil
}

// StashPath is GetStashSavePath (Source/pfile.cpp:94-104).
func (n SaveNamespace) StashPath(dir string) string {
	name := "stash"

	if n.Spawn {
		name = "stash_spawn"
	}

	return filepath.Join(dir, name+"."+n.extension())
}

// Set records a hero name in a slot (pfile.cpp:700-710).
func (s *HeroSlots) Set(slot int, name string) error {
	if slot < 0 || slot >= MaxCharacters {
		return fmt.Errorf("persist: hero slot %d out of range: %w", slot, ErrSlotOutOfRange)
	}

	s.names[slot] = name
	s.used[slot] = name != ""

	return nil
}

// Get returns the hero name of a slot, empty when the slot is free.
func (s *HeroSlots) Get(slot int) string {
	if slot < 0 || slot >= MaxCharacters {
		return ""
	}

	return s.names[slot]
}

// Count returns the number of occupied slots.
func (s *HeroSlots) Count() int {
	count := 0

	for _, used := range s.used {
		if used {
			count++
		}
	}

	return count
}

// Entries returns the occupied slots in save number order, the order
// pfile_ui_set_hero_infos walks the list and calls uiAddHeroInfo
// (Source/pfile.cpp:700-726).
func (s *HeroSlots) Entries() []HeroSlotEntry {
	entries := make([]HeroSlotEntry, 0, s.Count())

	for slot, used := range s.used {
		if used {
			entries = append(entries, HeroSlotEntry{Slot: slot, Name: s.names[slot]})
		}
	}

	return entries
}

// FirstFreeSlot returns the lowest free slot, or -1 when the list is full
// (MAX_CHARACTERS = 99 heroes, Source/pfile.h:24).
func (s *HeroSlots) FirstFreeSlot() int {
	for slot, used := range s.used {
		if !used {
			return slot
		}
	}

	return -1
}

// extension is the save extension, defaulting to "sv".
func (n SaveNamespace) extension() string {
	if n.Extension == "" {
		return "sv"
	}

	return n.Extension
}
