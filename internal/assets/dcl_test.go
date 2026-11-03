package assets_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// patchRtListfileStream is the decrypted PKWARE DCL stream of Patch_rt.mpq's
// "(listfile)" record (binary model, 1024-byte window). It is the sector
// payload minus the MPQ multi-method byte 0x08, taken from a record whose
// plaintext is the readable listfile below — so both sides are checkable by eye.
const patchRtListfileStream = "0004EAA4F912460E1D2E62DCB07973268D1B1750866069A0" +
	"90F68023E78D99346CCAE2BBCCA113864E9AB1DC0BCE18BC" +
	"64AF2F62D69CC5B51883268C9B3365E0849933E7CE1B3964" +
	"71ADDC94B91366CC983775DCD0C5052811232434A70CDB3A" +
	"73E316B1874C842672CA9CC5169A02A5C890244198E2C2A1" +
	"F3662EEE773789B5C51A0CCF9CB6ABE7ADA65D317AB89C2D" +
	"A0B74BDE1AA183876E5301FF"

// darkMageWalkStream is the decrypted PKWARE DCL stream of DIABDAT.MPQ's
// "monsters\darkmage\dmagew.cl2" sector (binary model, 4096-byte window).
const darkMageWalkStream = "000640000000E0E40F5CFEA0E60F78FE60E80F94FE20EA0F" +
	"730021343F02FE01"

// patchRtListfileText is the plaintext of Patch_rt.mpq's listfile record.
var patchRtListfileText = strings.Join([]string{
	`ui_art\bnlogin.PCX`,
	`ui_art\bnprofile.PCX`,
	`ui_art\bnstaticprofile.pcx`,
	`ui_art\bn_bkg.pcx`,
	`ui_art\changepassword.pcx`,
	`ui_art\newaccount.PCX`,
	`ui_art\profilebkg.pcx`,
	`ui_art\selconn.pcx`,
	`ui_art\selconn2.pcx`,
	`ui_art\selregn.pcx`,
	`ui_art\SPECIAL.PCX`,
	`ui_art\tos.PCX`,
	`ui_art\welcome.PCX`,
	`ui_art\xsmlogo.pcx`,
	`rez\gateways.txt`,
}, "\r\n") + "\r\n"

func TestExplodeDCLDecodesListfileSector(t *testing.T) {
	// given the decrypted Patch_rt.mpq (listfile) sector stream
	stream := decodeHex(t, patchRtListfileStream)

	// when the stream is exploded to its recorded 317-byte size
	decoded, err := assets.ExplodeDCL(stream, 317)
	// then the sector yields the listfile text verbatim
	if err != nil {
		t.Fatalf("ExplodeDCL: %v", err)
	}

	if got := string(decoded); got != patchRtListfileText {
		t.Errorf("listfile text:\n got %q\nwant %q", got, patchRtListfileText)
	}
}

func TestExplodeDCLDecodesWindow4096BinarySector(t *testing.T) {
	// given a decrypted DIABDAT sprite sector using a 4096-byte window
	stream := decodeHex(t, darkMageWalkStream)

	// when the stream is exploded to its recorded 96-byte size
	decoded, err := assets.ExplodeDCL(stream, 96)
	// then every decoded byte matches the record's plaintext
	if err != nil {
		t.Fatalf("ExplodeDCL: %v", err)
	}

	want := "20000000270000002e000000350000003c000000430000004a00000051000000" +
		"0000000008000000000000000800000000000000080000000000000008000000" +
		"0000000008000000000000000800000000000000080000000000000008000000"

	if got := hex.EncodeToString(decoded); got != want {
		t.Errorf("decoded bytes:\n got %s\nwant %s", got, want)
	}
}

func TestExplodeDCLStopsAtTheExpectedSize(t *testing.T) {
	// given a listfile stream whose plaintext is longer than the caller asked for
	stream := decodeHex(t, patchRtListfileStream)

	// when it is exploded to a shorter output size
	decoded, err := assets.ExplodeDCL(stream, 20)
	// then the output is truncated to that size instead of overflowing
	if err != nil {
		t.Fatalf("ExplodeDCL: %v", err)
	}

	if len(decoded) != 20 {
		t.Fatalf("decoded length = %d, want 20", len(decoded))
	}

	if got, want := string(decoded), patchRtListfileText[:20]; got != want {
		t.Errorf("decoded prefix = %q, want %q", got, want)
	}
}

func TestExplodeDCLRejectsShortStream(t *testing.T) {
	// given a stream too short to carry the two header bytes
	stream := []byte{0x00}

	// when it is exploded
	_, err := assets.ExplodeDCL(stream, 4)

	// then the decoder reports the missing header
	if err == nil {
		t.Fatal("ExplodeDCL accepted a 1-byte stream")
	}
}

func TestExplodeDCLRejectsASCIIModel(t *testing.T) {
	// given a stream declaring the ASCII literal model
	stream := decodeHex(t, "0104aabb")

	// when it is exploded
	_, err := assets.ExplodeDCL(stream, 4)

	// then the decoder names the unsupported model instead of guessing
	if err == nil {
		t.Fatal("ExplodeDCL accepted the ASCII literal model")
	}

	if !strings.Contains(err.Error(), "ascii") {
		t.Errorf("error = %v, want an ascii-model rejection", err)
	}
}

func TestExplodeDCLRejectsInvalidWindowBits(t *testing.T) {
	// given a stream declaring a three-bit window
	stream := decodeHex(t, "0003aabb")

	// when it is exploded
	_, err := assets.ExplodeDCL(stream, 4)

	// then the decoder reports the out-of-range window
	if err == nil {
		t.Fatal("ExplodeDCL accepted a 3-bit window")
	}

	if !strings.Contains(err.Error(), "window") {
		t.Errorf("error = %v, want a window-size rejection", err)
	}
}

func TestExplodeDCLRejectsTruncatedStream(t *testing.T) {
	// given the first eight bytes of a valid stream
	stream := decodeHex(t, darkMageWalkStream)[:8]

	// when it is exploded to the size the full stream carried
	_, err := assets.ExplodeDCL(stream, 96)

	// then the decoder reports the truncation rather than returning short data
	if err == nil {
		t.Fatal("ExplodeDCL accepted a truncated stream")
	}
}

func TestExplodeDCLRejectsBackReferenceBeforeStart(t *testing.T) {
	// given a stream whose first symbol is a match, built by hand from the
	// length/distance tables: flag 1, length code 0x01 (3 bits) = length 4,
	// distance code 0x03 (2 bits) with a zero distance
	stream := decodeHex(t, "00063900")

	// when it is exploded
	_, err := assets.ExplodeDCL(stream, 4)

	// then the decoder reports the distance that reaches before the output
	if err == nil {
		t.Fatal("ExplodeDCL accepted a match before the first literal")
	}

	if !strings.Contains(err.Error(), "distance") {
		t.Errorf("error = %v, want a distance rejection", err)
	}
}

// decodeHex parses a hex literal fixture, failing the test on a typo.
func decodeHex(t *testing.T, literal string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(literal)
	if err != nil {
		t.Fatalf("hex literal: %v", err)
	}

	return decoded
}
