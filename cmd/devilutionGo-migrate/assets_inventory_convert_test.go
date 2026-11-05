package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetsConvertWritesTheGoldensTheConvertersProduce(t *testing.T) {
	// given a committed CEL input and the C++ oracle's CLX output for it
	input := fixtureBytes(t, "cel/data_talkbutton.cel.hex")
	want := fixtureBytes(t, "cel/data_talkbutton.cel.clx.hex")
	inputPath := filepath.Join(t.TempDir(), "talkbutton.cel")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	outPath := filepath.Join(t.TempDir(), "talkbutton.clx")

	// when the nested convert subcommand runs
	code, stdout, stderr := runMigrate(t, "assets", "convert",
		"--format", "cel", "--in", inputPath, "--out", outPath, "--widths", "53,53")

	// then the written bytes match the oracle exactly
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, stderr)
	}

	if !bytes.Contains([]byte(stdout), []byte("wrote")) {
		t.Fatalf("stdout = %q, want the written-bytes report", stdout)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("converted %d bytes, want %d", len(got), len(want))
	}
}
