package assets_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mpq "github.com/suprsokr/go-mpq"

	"github.com/bmcszk/devilutionGo/internal/engine/assets"
)

// innerName mirrors a real MPQ inner path; MPQ convention uses backslashes.
const innerName = `txtdata\monsters.tsv`

func TestStackPriorityOverride(t *testing.T) {
	// given one low and one high priority archive holding the same asset
	lowDir := filepath.Join(t.TempDir(), "low")
	highDir := filepath.Join(t.TempDir(), "high")
	writeMpq(t, lowDir, "spawn", "low-spawn")
	writeMpq(t, highDir, "spawn", "high-spawn")

	// when three stacks load the pair in each scenario's order
	highSecond := &assets.Stack{}
	loadOK(t, highSecond, lowDir, assets.MainMpqPriority)
	loadOK(t, highSecond, highDir, 9000)

	highFirst := &assets.Stack{}
	loadOK(t, highFirst, highDir, assets.DevilutionXMpqPriority)
	loadOK(t, highFirst, lowDir, assets.MainMpqPriority)

	logs := captureLog(t)
	collision := assets.Open([]string{lowDir}, "spawn")
	defer collision.Close()

	collisionLoaded := collision.Load([]string{highDir}, "spawn", assets.MainMpqPriority)
	collisionData, collisionOK := collision.ReadFile(innerName)

	// then the higher priority archive wins or the existing archive is kept
	t.Run("higher priority wins", func(t *testing.T) {
		assertInnerRead(t, highSecond, "high-spawn")
	})

	t.Run("priority wins regardless of load order", func(t *testing.T) {
		assertInnerRead(t, highFirst, "high-spawn")
	})

	t.Run("priority collision keeps existing archive", func(t *testing.T) {
		if !collisionLoaded {
			t.Fatal("colliding Load returned false")
		}

		if !collisionOK || string(collisionData) != "low-spawn" {
			t.Fatalf("ReadFile = %q, %v; want low-spawn (existing kept)", collisionData, collisionOK)
		}

		if !strings.Contains(logs.String(), "already registered") {
			t.Fatalf("logs = %q; want collision warning", logs.String())
		}
	})
}

func TestStackUnpackedDir(t *testing.T) {
	// given an unpacked spawn tree and a packed archive holding the same asset
	dir := t.TempDir()
	unpacked := filepath.Join(dir, "spawn")
	writeUnpackedDir(t, unpacked, "unpacked-spawn")

	mpqDir := filepath.Join(dir, "mpqs")
	writeMpq(t, mpqDir, "spawn", "packed-spawn")

	// when one stack reads the unpacked tree and another loads archive first
	unpackedOnly := &assets.Stack{}
	loadUnpackedOK(t, unpackedOnly, dir, assets.MainMpqPriority)
	backslash, backslashOK := unpackedOnly.ReadFile(innerName)
	forward, forwardOK := unpackedOnly.ReadFile("txtdata/monsters.tsv")

	unpackedWins := &assets.Stack{}
	loadOK(t, unpackedWins, mpqDir, assets.MainMpqPriority)
	loadUnpackedOK(t, unpackedWins, dir, 9000)

	// then both spellings reach the unpacked tree and it outranks the archive
	t.Run("backslash name reads forward-slash tree", func(t *testing.T) {
		if !backslashOK || string(backslash) != "unpacked-spawn" {
			t.Fatalf("ReadFile backslash = %q, %v", backslash, backslashOK)
		}

		if !forwardOK || string(forward) != "unpacked-spawn" {
			t.Fatalf("ReadFile forward slash = %q, %v", forward, forwardOK)
		}
	})

	t.Run("unpacked dir outranks lower priority archive", func(t *testing.T) {
		assertInnerRead(t, unpackedWins, "unpacked-spawn")
	})
}

func TestStackMissingAssets(t *testing.T) {
	// given
	stack := assets.Open([]string{filepath.Join(t.TempDir(), "absent")}, "spawn")

	// when
	got, gotOK := stack.ReadFile(innerName)
	unpackedOK := stack.LoadUnpacked([]string{filepath.Join(t.TempDir(), "absent")}, "spawn", 1)
	_, emptyNameOK := stack.ReadFile("")
	_, trailingSlashOK := stack.ReadFile(`txtdata\`)

	// then
	if gotOK {
		t.Fatalf("ReadFile on empty stack = %q; want miss", got)
	}

	if unpackedOK {
		t.Fatal("LoadUnpacked found a directory that does not exist")
	}

	if emptyNameOK {
		t.Fatal("empty asset name resolved")
	}

	if trailingSlashOK {
		t.Fatal("trailing-backslash asset name resolved")
	}
}

// Reference: diasurgical/DevilutionX test/timedemo_test.cpp:47-49 — real-asset
// tests skip politely when no game archive is installed.
func TestStackRealMpqWhenPresent(t *testing.T) {
	// given
	path := os.Getenv("DG_TEST_MPQ")
	if path == "" {
		t.Skip("MPQ assets (spawn.mpq or DIABDAT.MPQ) not found - skipping test")
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	stack := assets.Open([]string{filepath.Dir(path)}, name)
	defer stack.Close()

	// when
	data, ok := stack.ReadFile(innerName)

	// then
	if !ok || len(data) == 0 {
		t.Fatalf("real archive %s did not serve %q", path, innerName)
	}
}

// loadOK loads one archive into the stack or fails the test.
func loadOK(t *testing.T, stack *assets.Stack, dir string, prio int) {
	t.Helper()

	if !stack.Load([]string{dir}, "spawn", prio) {
		t.Fatalf("archive %s did not load", dir)
	}
}

// loadUnpackedOK loads one unpacked dir into the stack or fails the test.
func loadUnpackedOK(t *testing.T, stack *assets.Stack, dir string, prio int) {
	t.Helper()

	if !stack.LoadUnpacked([]string{dir}, "spawn", prio) {
		t.Fatalf("unpacked dir %s did not load", dir)
	}
}

// assertInnerRead reads innerName and pins its content.
func assertInnerRead(t *testing.T, stack *assets.Stack, want string) {
	t.Helper()

	got, ok := stack.ReadFile(innerName)
	if !ok || string(got) != want {
		t.Fatalf("ReadFile = %q, %v; want %s", got, ok, want)
	}
}

func writeMpq(t *testing.T, root, name, content string) {
	t.Helper()

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(root, name+"-src.txt")
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	archive, err := mpq.Create(filepath.Join(root, name+".mpq"), 1)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}

	if err := archive.AddFile(src, innerName); err != nil {
		t.Fatalf("add %s: %v", innerName, err)
	}

	if err := archive.Close(); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func writeUnpackedDir(t *testing.T, root, content string) {
	t.Helper()

	dir := filepath.Join(root, "txtdata")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "monsters.tsv"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	previousOut := log.Writer()
	previousFlags := log.Flags()

	log.SetOutput(buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(previousOut)
		log.SetFlags(previousFlags)
	})

	return buf
}
