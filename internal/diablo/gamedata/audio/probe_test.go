package audio_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// The E-A1 probe: no retail sound byte had ever been decoded in Go, and
// docs/go-libraries.md asserted "Retail = 22050 Hz 8-bit unsigned PCM WAV"
// WITHOUT a cited measurement (go-libraries.md:95). These tests answer that with
// real records: they read the containers through the same MPQStack the boot
// uses, parse the RIFF header by hand, and record the measured fmt chunk.
//
// The assertion is the wave's gate: ebiten's wav decoder accepts ONLY linear
// PCM, 1 or 2 channels, 8 or 16 bits (audio/wav/decode.go:187-203), so a
// non-PCM tag would STOP the audio leg.
//
// The retail archives are environment inputs, never committed fixtures, so an
// absent DG_ASSET_ARCHIVES skips (mpq_smoke_test.go findAssetArchives
// politeness). A PRESENT archive whose sound is not PCM is a FAIL.

// waveFormatPCM is WAVE_FORMAT_PCM (mmreg.h), the only tag ebiten's decoder
// takes (audio/wav/decode.go:187-189).
const waveFormatPCM = 1

// probeSfxRecords are the sfx containers the gate parses: the catalog's first
// row (effects.tsv "Walk  Misc  sfx\\misc\\walk1.wav") plus its neighbour.
var probeSfxRecords = []string{"sfx/misc/walk1.wav", "sfx/misc/walk2.wav"}

// probeMusicRecords are level tracks MusicTracks names
// (Source/engine/sound.cpp:158-167). Retail music ships as a separate add-on
// for Hellfire (hfmusic.mpq), so an install without it legitimately lacks some.
var probeMusicRecords = []string{"music/dlvlf.wav", "music/dtowne.wav", "music/dintro.wav"}

// waveHeader is the measured RIFF "fmt " chunk of one retail sound.
type waveHeader struct {
	// FormatTag is wFormatTag (1 = PCM, 2 = MS ADPCM, 0x11 = IMA ADPCM).
	FormatTag uint16
	// Channels is the channel count.
	Channels uint16
	// SampleRate is the sample rate in Hz.
	SampleRate uint32
	// ByteRate is the average bytes per second.
	ByteRate uint32
	// BlockAlign is the block alignment in bytes.
	BlockAlign uint16
	// BitsPerSample is the sample width in bits.
	BitsPerSample uint16
	// DataSize is the RIFF "data" chunk length.
	DataSize uint32
	// RIFFSize is the declared RIFF chunk size.
	RIFFSize uint32
}

// waveChunk is one RIFF chunk body.
type waveChunk struct {
	id   string
	body []byte
}

func TestSfxProbeRetailWaveFormat(t *testing.T) {
	stack := openRetailStack(t)

	defer func() { _ = stack.Close() }()

	for _, assetName := range probeSfxRecords {
		t.Run(assetName, func(t *testing.T) {
			// when parsing a REQUIRED sfx record's RIFF header
			header := probeRequiredWave(t, stack, assetName)

			// then the payload is linear PCM
			assertLinearPCM(t, assetName, header)
		})
	}
}

func TestMusicProbeRetailWaveFormat(t *testing.T) {
	stack := openRetailStack(t)

	defer func() { _ = stack.Close() }()

	for _, assetName := range probeMusicRecords {
		t.Run(assetName, func(t *testing.T) {
			// when parsing an OPTIONAL music record's RIFF header
			header := probeOptionalWave(t, stack, assetName)
			if header == nil {
				return
			}

			// then a present payload is linear PCM too
			assertLinearPCM(t, assetName, *header)
		})
	}
}

// String renders the fmt chunk for the measurement log.
func (h waveHeader) String() string {
	return fmt.Sprintf("fmt tag=%d (%s) channels=%d rate=%d Hz byteRate=%d blockAlign=%d bits=%d "+
		"data=%d bytes riff=%d bytes",
		h.FormatTag, codecName(h.FormatTag), h.Channels, h.SampleRate, h.ByteRate, h.BlockAlign,
		h.BitsPerSample, h.DataSize, h.RIFFSize)
}

// codecName names the WAVE format tag.
func codecName(tag uint16) string {
	switch tag {
	case waveFormatPCM:
		return "PCM"
	case 2:
		return "MS ADPCM"
	case 0x11:
		return "IMA ADPCM"
	default:
		return "unknown"
	}
}

// openRetailStack loads the retail archives named by DG_ASSET_ARCHIVES, or
// skips the test when none is configured.
func openRetailStack(t *testing.T) *assets.MPQStack {
	t.Helper()

	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	stack := assets.NewMPQStack()
	stack.Load(dirs, "hellfire", assets.HellfireMpqPriority)
	stack.Load(dirs, "DIABDAT", assets.MainMpqPriority, ".MPQ")

	return stack
}

// probeRequiredWave parses a record that MUST be present.
func probeRequiredWave(t *testing.T, stack *assets.MPQStack, assetName string) waveHeader {
	t.Helper()

	data, ok := stack.ReadFile(assetName)
	if !ok || len(data) == 0 {
		t.Fatalf("%s is not present in the archives", assetName)
	}

	return parseMeasuredWave(t, assetName, data)
}

// probeOptionalWave parses a record that may be absent, reporting nil then.
func probeOptionalWave(t *testing.T, stack *assets.MPQStack, assetName string) *waveHeader {
	t.Helper()

	data, ok := stack.ReadFile(assetName)
	if !ok || len(data) == 0 {
		t.Logf("ABSENT %s (not in these archives)", assetName)

		return nil
	}

	header := parseMeasuredWave(t, assetName, data)

	return &header
}

// parseMeasuredWave parses one container and logs the measured fmt chunk.
func parseMeasuredWave(t *testing.T, assetName string, data []byte) waveHeader {
	t.Helper()

	header, err := parseWaveHeader(data)
	if err != nil {
		t.Fatalf("%s: %v", assetName, err)
	}

	t.Logf("MEASURED %s: %s", assetName, header)

	return header
}

// assertLinearPCM fails when the measured container is not linear PCM.
func assertLinearPCM(t *testing.T, assetName string, header waveHeader) {
	t.Helper()

	if header.FormatTag != waveFormatPCM {
		t.Fatalf("%s: format tag %d (0x%04x, %s), want %d (linear PCM): STOP the audio leg",
			assetName, header.FormatTag, header.FormatTag, codecName(header.FormatTag), waveFormatPCM)
	}
}

// parseWaveHeader reads the RIFF container and returns its fmt chunk plus the
// data length (the same walk ebiten's decoder does, audio/wav/decode.go:150-215).
func parseWaveHeader(data []byte) (waveHeader, error) {
	if err := checkRIFF(data); err != nil {
		return waveHeader{}, err
	}

	chunks, err := riffChunks(data)
	if err != nil {
		return waveHeader{}, err
	}

	header, err := findFmtHeader(chunks)
	if err != nil {
		return waveHeader{}, err
	}

	if header.DataSize, err = findDataSize(chunks); err != nil {
		return waveHeader{}, err
	}

	header.RIFFSize = binary.LittleEndian.Uint32(data[4:8])

	return header, nil
}

// checkRIFF validates the 12-byte RIFF/WAVE preamble.
func checkRIFF(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("wave: %d bytes is shorter than a RIFF header", len(data))
	}

	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fmt.Errorf("wave: not a RIFF/WAVE container (%.4s/%.4s)", data[0:4], data[8:12])
	}

	return nil
}

// riffChunks splits the container's chunk list, honouring the word-alignment
// pad byte after an odd-sized chunk.
func riffChunks(data []byte) ([]waveChunk, error) {
	var chunks []waveChunk

	for offset := 12; offset+8 <= len(data); {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		body := offset + 8

		if body+size > len(data) {
			return nil, fmt.Errorf("wave: chunk %.4s declares %d bytes past the end", id, size)
		}

		chunks = append(chunks, waveChunk{id: id, body: data[body : body+size]})
		offset = body + size + size%2
	}

	return chunks, nil
}

// findFmtHeader reads the fmt chunk's fields.
func findFmtHeader(chunks []waveChunk) (waveHeader, error) {
	for _, chunk := range chunks {
		if chunk.id == "fmt " {
			return parseFmtChunk(chunk.body)
		}
	}

	return waveHeader{}, errors.New("wave: no fmt chunk found")
}

// parseFmtChunk decodes the 16-byte PCM WAVEFORMAT fields.
func parseFmtChunk(body []byte) (waveHeader, error) {
	if len(body) < 16 {
		return waveHeader{}, fmt.Errorf("wave: fmt chunk is %d bytes, want >= 16", len(body))
	}

	return waveHeader{
		FormatTag:     binary.LittleEndian.Uint16(body[0:2]),
		Channels:      binary.LittleEndian.Uint16(body[2:4]),
		SampleRate:    binary.LittleEndian.Uint32(body[4:8]),
		ByteRate:      binary.LittleEndian.Uint32(body[8:12]),
		BlockAlign:    binary.LittleEndian.Uint16(body[12:14]),
		BitsPerSample: binary.LittleEndian.Uint16(body[14:16]),
	}, nil
}

// findDataSize returns the data chunk's length.
func findDataSize(chunks []waveChunk) (uint32, error) {
	for _, chunk := range chunks {
		if chunk.id == "data" {
			return uint32(len(chunk.body)), nil
		}
	}

	return 0, errors.New("wave: no data chunk found")
}
