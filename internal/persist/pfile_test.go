package persist_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestHeroPath(t *testing.T) {
	// given each save namespace (pfile.cpp:79-92)
	cases := []struct {
		name string
		ns   persist.SaveNamespace
		slot int
		want string
	}{
		{"single", persist.SaveNamespace{}, 0, "dir/single_0.sv"},
		{"multi", persist.SaveNamespace{Multiplayer: true}, 3, "dir/multi_3.sv"},
		{"spawn", persist.SaveNamespace{Spawn: true}, 1, "dir/spawn_1.sv"},
		{"spawn multi", persist.SaveNamespace{Spawn: true, Multiplayer: true}, 2, "dir/share_2.sv"},
		{"mod extension", persist.SaveNamespace{Extension: "svx"}, 4, "dir/single_4.svx"},
	}
	for _, tc := range cases {
		// when
		got, err := tc.ns.HeroPath("dir", tc.slot)
		// then the file name matches pfile's assembly
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}

		if got != tc.want {
			t.Fatalf("%s = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSavePathWithPrefix(t *testing.T) {
	// when a demo save prefix is used (pfile.cpp:661-663)
	got, err := persist.SaveNamespace{}.SavePath("dir", 5, "demo_1_reference_")
	// then the prefix lands before the namespace token
	if err != nil {
		t.Fatalf("SavePath: %v", err)
	}

	if got != "dir/demo_1_reference_single_5.sv" {
		t.Fatalf("SavePath = %q", got)
	}
}

func TestSavePathSlotOutOfRange(t *testing.T) {
	// when the slot reaches MAX_CHARACTERS (pfile.h:24)
	_, err := persist.SaveNamespace{}.HeroPath("dir", persist.MaxCharacters)
	// then the path is refused
	if !errors.Is(err, persist.ErrSlotOutOfRange) {
		t.Fatalf("err = %v, want ErrSlotOutOfRange", err)
	}
}

func TestStashPath(t *testing.T) {
	// given the two stash families (pfile.cpp:94-104)
	namespaces := []persist.SaveNamespace{{}, {Spawn: true}}
	// when
	got := []string{namespaces[0].StashPath("dir"), namespaces[1].StashPath("dir")}
	// then each namespace gets its own stash file
	want := []string{"dir/stash.sv", "dir/stash_spawn.sv"}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("StashPath = %q, want %q", got[i], want[i])
		}
	}
}

func TestHeroSlots(t *testing.T) {
	// given an empty hero list
	var slots persist.HeroSlots
	// when three heroes are recorded
	for i, name := range []string{"Conan", "Rogue", "Aidan"} {
		if err := slots.Set(i, name); err != nil {
			t.Fatalf("Set(%d): %v", i, err)
		}
	}
	// then the count, lookup and first free slot follow the list
	if slots.Count() != 3 {
		t.Fatalf("Count = %d, want 3", slots.Count())
	}

	if slots.Get(1) != "Rogue" {
		t.Fatalf("Get(1) = %q", slots.Get(1))
	}

	if slots.FirstFreeSlot() != 3 {
		t.Fatalf("FirstFreeSlot = %d, want 3", slots.FirstFreeSlot())
	}
}

func TestHeroSlotsFull(t *testing.T) {
	// given all 99 slots taken
	var slots persist.HeroSlots
	for i := range persist.MaxCharacters {
		if err := slots.Set(i, "Hero"); err != nil {
			t.Fatalf("Set(%d): %v", i, err)
		}
	}
	// when a free slot is requested
	first := slots.FirstFreeSlot()
	// then the list reports none and slot 99 is out of range
	if first != -1 {
		t.Fatalf("FirstFreeSlot = %d, want -1", first)
	}

	if err := slots.Set(persist.MaxCharacters, "Overflow"); !errors.Is(err, persist.ErrSlotOutOfRange) {
		t.Fatalf("err = %v, want ErrSlotOutOfRange", err)
	}
}
