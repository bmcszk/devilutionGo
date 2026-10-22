package assets_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/assets"
	mpq "github.com/suprsokr/go-mpq"
)

// innerName mirrors a real MPQ inner path; MPQ convention uses backslashes.
const innerName = `txtdata\monsters.tsv`

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

func TestStackPriorityOverride(t *testing.T) {
	// given
	lowDir := filepath.Join(t.TempDir(), "low")
	highDir := filepath.Join(t.TempDir(), "high")
	writeMpq(t, lowDir, "spawn", "low-spawn")
	writeMpq(t, highDir, "spawn", "high-spawn")

	t.Run("higher priority wins regardless of load order", func(t *testing.T) {
		lowThenHigh := &assets.Stack{}
		if !lowThenHigh.Load([]string{lowDir}, "spawn", assets.MainMpqPriority) {
			t.Fatal("low archive did not load")
		}
		if !lowThenHigh.Load([]string{highDir}, "spawn", 9000) {
			t.Fatal("high archive did not load")
		}
		got, ok := lowThenHigh.ReadFile(innerName)
		if !ok || string(got) != "high-spawn" {
			t.Fatalf("ReadFile = %q, %v; want high-spawn", got, ok)
		}

		highThenLow := &assets.Stack{}
		if !highThenLow.Load([]string{highDir}, "spawn", assets.DevilutionXMpqPriority) {
			t.Fatal("high archive did not load")
		}
		if !highThenLow.Load([]string{lowDir}, "spawn", assets.MainMpqPriority) {
			t.Fatal("low archive did not load")
		}
		got, ok = highThenLow.ReadFile(innerName)
		if !ok || string(got) != "high-spawn" {
			t.Fatalf("ReadFile = %q, %v; want high-spawn", got, ok)
		}
	})

	t.Run("priority collision keeps existing archive", func(t *testing.T) {
		logs := captureLog(t)
		stack := assets.Open([]string{lowDir}, "spawn")
		defer stack.Close()
		if !stack.Load([]string{highDir}, "spawn", assets.MainMpqPriority) {
			t.Fatal("colliding Load returned false")
		}
		got, ok := stack.ReadFile(innerName)
		if !ok || string(got) != "low-spawn" {
			t.Fatalf("ReadFile = %q, %v; want low-spawn (existing kept)", got, ok)
		}
		if !strings.Contains(logs.String(), "already registered") {
			t.Fatalf("logs = %q; want collision warning", logs.String())
		}
	})
}

func TestStackUnpackedDir(t *testing.T) {
	// given
	dir := t.TempDir()
	unpacked := filepath.Join(dir, "spawn")
	writeUnpackedDir(t, unpacked, "unpacked-spawn")
	mpqDir := filepath.Join(dir, "mpqs")
	writeMpq(t, mpqDir, "spawn", "packed-spawn")

	t.Run("backslash name reads forward-slash tree", func(t *testing.T) {
		stack := &assets.Stack{}
		if !stack.LoadUnpacked([]string{dir}, "spawn", assets.MainMpqPriority) {
			t.Fatal("unpacked dir did not load")
		}
		got, ok := stack.ReadFile(innerName)
		if !ok || string(got) != "unpacked-spawn" {
			t.Fatalf("ReadFile backslash = %q, %v", got, ok)
		}
		got, ok = stack.ReadFile("txtdata/monsters.tsv")
		if !ok || string(got) != "unpacked-spawn" {
			t.Fatalf("ReadFile forward slash = %q, %v", got, ok)
		}
	})

	t.Run("unpacked dir outranks lower priority archive", func(t *testing.T) {
		stack := &assets.Stack{}
		if !stack.Load([]string{mpqDir}, "spawn", assets.MainMpqPriority) {
			t.Fatal("archive did not load")
		}
		if !stack.LoadUnpacked([]string{dir}, "spawn", 9000) {
			t.Fatal("unpacked dir did not load")
		}
		got, ok := stack.ReadFile(innerName)
		if !ok || string(got) != "unpacked-spawn" {
			t.Fatalf("ReadFile = %q, %v; want unpacked-spawn", got, ok)
		}
	})
}

func TestStackMissingAssets(t *testing.T) {
	// given
	stack := assets.Open([]string{filepath.Join(t.TempDir(), "absent")}, "spawn")
	if got, ok := stack.ReadFile(innerName); ok {
		t.Fatalf("ReadFile on empty stack = %q, %v; want miss", got, ok)
	}
	if stack.LoadUnpacked([]string{filepath.Join(t.TempDir(), "absent")}, "spawn", 1) {
		t.Fatal("LoadUnpacked found a directory that does not exist")
	}
	if _, ok := stack.ReadFile(""); ok {
		t.Fatal("empty asset name resolved")
	}
	if _, ok := stack.ReadFile(`txtdata\`); ok {
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
	if data, ok := stack.ReadFile(innerName); !ok || len(data) == 0 {
		t.Fatalf("real archive %s did not serve %q", path, innerName)
	}
}
