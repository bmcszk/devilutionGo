package assets_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	mpq "github.com/suprsokr/go-mpq"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// innerName mirrors a real MPQ inner path; the archive convention uses backslashes.
const innerName = `txtdata\monsters.tsv`

func TestMPQStackPriorityOverride(t *testing.T) {
	// given a low and a high priority archive holding the same asset
	lowDir := filepath.Join(t.TempDir(), "low")
	highDir := filepath.Join(t.TempDir(), "high")
	writeMPQ(t, lowDir, map[string]string{innerName: "low-spawn"})
	writeMPQ(t, highDir, map[string]string{innerName: "high-spawn"})

	// when two stacks load the pair in opposite orders
	highSecond := assets.NewMPQStack()
	closeOnCleanup(t, highSecond)
	loadArchive(t, highSecond, lowDir, assets.MainMpqPriority)
	loadArchive(t, highSecond, highDir, assets.DevilutionXMpqPriority)

	highFirst := assets.NewMPQStack()
	closeOnCleanup(t, highFirst)
	loadArchive(t, highFirst, highDir, assets.DevilutionXMpqPriority)
	loadArchive(t, highFirst, lowDir, assets.MainMpqPriority)

	// then the higher priority archive serves the asset either way
	t.Run("higher priority loaded second wins", func(t *testing.T) {
		assertInnerRead(t, highSecond, "high-spawn")
	})

	t.Run("higher priority loaded first wins", func(t *testing.T) {
		assertInnerRead(t, highFirst, "high-spawn")
	})
}

func TestMPQStackDuplicatePriority(t *testing.T) {
	// given two archives that will register at the same priority
	firstDir := filepath.Join(t.TempDir(), "first")
	secondDir := filepath.Join(t.TempDir(), "second")
	writeMPQ(t, firstDir, map[string]string{innerName: "first-spawn"})
	writeMPQ(t, secondDir, map[string]string{innerName: "second-spawn"})

	logs := captureSlog(t)
	stack := assets.NewMPQStack()
	closeOnCleanup(t, stack)

	// when both load at MainMpqPriority
	firstLoaded := stack.Load([]string{firstDir}, "spawn", assets.MainMpqPriority)
	secondLoaded := stack.Load([]string{secondDir}, "spawn", assets.MainMpqPriority)

	// then the collision keeps the first archive and warns
	if !firstLoaded || !secondLoaded {
		t.Fatalf("Load = %v, %v; want both true", firstLoaded, secondLoaded)
	}

	assertInnerRead(t, stack, "first-spawn")

	if !strings.Contains(logs.String(), "already registered") {
		t.Fatalf("logs = %q; want an already registered warning", logs.String())
	}
}

func TestMPQStackMissingArchive(t *testing.T) {
	// given a stack and a search path that holds no archive
	stack := assets.NewMPQStack()
	closeOnCleanup(t, stack)

	// when Load probes the empty path
	loaded := stack.Load([]string{filepath.Join(t.TempDir(), "absent")}, "spawn", assets.MainMpqPriority)
	data, dataOK := stack.ReadFile(innerName)

	// then the miss is reported and nothing is served
	if loaded {
		t.Fatal("Load returned true for a search path without the archive")
	}

	if dataOK || data != nil {
		t.Fatalf("ReadFile = %q, %v; want a miss", data, dataOK)
	}
}

func TestMPQStackUnpackedDirectory(t *testing.T) {
	// given an unpacked extraction tree and a packed archive holding the same asset
	root := t.TempDir()
	writeUnpacked(t, filepath.Join(root, "spawn"), map[string]string{innerName: "unpacked-spawn"})

	mpqDir := filepath.Join(root, "mpqs")
	writeMPQ(t, mpqDir, map[string]string{innerName: "packed-spawn"})

	// when one stack loads only the tree and another prefers it to the archive
	unpackedOnly := assets.NewMPQStack()
	closeOnCleanup(t, unpackedOnly)
	loadUnpacked(t, unpackedOnly, root, "spawn", assets.MainMpqPriority)

	unpackedWins := assets.NewMPQStack()
	closeOnCleanup(t, unpackedWins)
	loadArchive(t, unpackedWins, mpqDir, assets.MainMpqPriority)
	loadUnpacked(t, unpackedWins, root, "spawn", assets.DevilutionXMpqPriority)

	// then both separators reach the tree and it outranks the archive
	t.Run("both separators reach the unpacked tree", func(t *testing.T) {
		assertInnerRead(t, unpackedOnly, "unpacked-spawn")

		forward, forwardOK := unpackedOnly.ReadFile("txtdata/monsters.tsv")
		if !forwardOK || string(forward) != "unpacked-spawn" {
			t.Fatalf("ReadFile forward slash = %q, %v; want unpacked-spawn", forward, forwardOK)
		}
	})

	t.Run("unpacked tree outranks the lower priority archive", func(t *testing.T) {
		assertInnerRead(t, unpackedWins, "unpacked-spawn")
	})
}

func TestMPQStackSatisfiesAssetSource(t *testing.T) {
	// when the stack is assigned to the AssetSource seam
	var _ assets.AssetSource = assets.NewMPQStack()
	var source assets.AssetSource = assets.NewMPQStack()
	data, dataOK := source.ReadFile(innerName)

	// then the assignment compiled and an empty stack serves nothing
	if dataOK || data != nil {
		t.Fatalf("ReadFile = %q, %v; want a miss", data, dataOK)
	}
}

func TestMPQStackFiles(t *testing.T) {
	// given a low priority archive with a differently spelled shared asset and
	// a high priority archive with an asset of its own
	highDir := filepath.Join(t.TempDir(), "high")
	lowDir := filepath.Join(t.TempDir(), "low")
	writeMPQ(t, highDir, map[string]string{
		innerName:           "high-monsters",
		`txtdata\items.tsv`: "high-items",
	})
	writeMPQ(t, lowDir, map[string]string{
		`TxtData\Monsters.tsv`: "low-monsters",
		`txtdata\spells.tsv`:   "low-spells",
	})

	// when the stack loads low first and high second
	stack := assets.NewMPQStack()
	closeOnCleanup(t, stack)
	loadArchive(t, stack, lowDir, assets.MainMpqPriority)
	loadArchive(t, stack, highDir, assets.DevilutionXMpqPriority)

	// then Files dedupes case-insensitively, names the entry after the
	// higher-priority archive and lists everything the stack serves
	want := []string{"txtdata/items.tsv", "txtdata/monsters.tsv", "txtdata/spells.tsv"}
	if got := stack.Files(); !slices.Equal(got, want) {
		t.Fatalf("Files = %q; want %q", got, want)
	}
}

// loadArchive loads one archive into the stack or fails the test.
func loadArchive(t *testing.T, stack *assets.MPQStack, dir string, priority int) {
	t.Helper()

	if !stack.Load([]string{dir}, "spawn", priority) {
		t.Fatalf("archive %s at priority %d did not load", dir, priority)
	}
}

// loadUnpacked registers one unpacked tree in the stack or fails the test.
func loadUnpacked(t *testing.T, stack *assets.MPQStack, dir, name string, priority int) {
	t.Helper()

	if !stack.LoadUnpacked([]string{dir}, name, priority) {
		t.Fatalf("unpacked %s at priority %d did not load", name, priority)
	}
}

// assertInnerRead reads innerName and pins its content.
func assertInnerRead(t *testing.T, stack *assets.MPQStack, want string) {
	t.Helper()

	got, gotOK := stack.ReadFile(innerName)
	if !gotOK || string(got) != want {
		t.Fatalf("ReadFile = %q, %v; want %s", got, gotOK, want)
	}
}

// writeMPQ builds <root>/spawn.mpq, the archive name every stack test loads,
// holding the given inner name to content map.
func writeMPQ(t *testing.T, root string, files map[string]string) {
	t.Helper()

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	archive, err := mpq.Create(filepath.Join(root, "spawn.mpq"), 1)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}

	addMPQFiles(t, archive, root, files)

	if err := archive.Close(); err != nil {
		t.Fatalf("write archive: %v", err)
	}
}

// addMPQFiles writes each content to a source file and adds it under its inner name.
func addMPQFiles(t *testing.T, archive *mpq.Archive, root string, files map[string]string) {
	t.Helper()

	for inner, content := range files {
		src := filepath.Join(root, strings.ReplaceAll(inner, `\`, "_")+"-src.txt")
		if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := archive.AddFile(src, inner); err != nil {
			t.Fatalf("add %s: %v", inner, err)
		}
	}
}

// closeOnCleanup closes the stack when the test ends.
func closeOnCleanup(t *testing.T, stack *assets.MPQStack) {
	t.Helper()

	t.Cleanup(func() {
		if err := stack.Close(); err != nil {
			t.Errorf("close stack: %v", err)
		}
	})
}

// writeUnpacked lays out an unpacked extraction tree holding the given inner
// name to content map.
func writeUnpacked(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for inner, content := range files {
		path := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(inner, `\`, "/")))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// captureSlog redirects the default slog output into a buffer for the duration
// of the test.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return buf
}
