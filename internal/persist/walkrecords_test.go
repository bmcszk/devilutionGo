package persist_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// lightingWantHex is the pin image of a standing (vision=false) light:
// tile (12,34), radius 9, isInvalid 0, hasChanged 1, old (11,35), oldRadius 8,
// offset (-1,2), vision flag 0 (loadsave.cpp:1795-1812).
const lightingWantHex = "0c000000220000000900000000000000" +
	"0000000001000000000000000b00000023000000" +
	"08000000ffffffff0200000000000000"

// portalWantHex is the pin image of a closed portal pointing at level 4 with
// ltype 1 (loadsave.cpp:1813-1822): open=0, position (75,68), level=4,
// ltype=1, setlvl=0.
const portalWantHex = "000000004b00000044000000" +
	"040000000100000000000000"

// objectWantHex is the pin image of a static torch record with the dynamic
// light var1 zeroed (loadsave.cpp:1703-1746): otype OBJ_L1LIGHT (1157),
// position (63, 62), applyLighting 1, anim flag 0, delay 5, cnt 0, len 10,
// frame 1, width 96, width2 (sprite tile center) 0, delFlag 0, break 0,
// solid 1, miss 0, selFlag 4, pre 0, trap 0, door 0, olid 3, rndSeed 42,
// var1 zeroed by the dynamic-light list (was -1), vars 2-6 zero, bookMessage
// 0, var8 0.
const objectWantHex = "85040000" + // _otype OBJ_L1LIGHT
	"3f000000" + // position.x
	"3e000000" + // position.y
	"01000000" + // applyLighting
	"00000000" + // _oAnimFlag
	"00000000" + // _oAnimData pointer skip
	"05000000" + // _oAnimDelay
	"00000000" + // _oAnimCnt
	"0a000000" + // _oAnimLen
	"01000000" + // _oAnimFrame
	"60000000" + // _oAnimWidth
	"00000000" + // _oAnimWidth2
	"00000000" + // _oDelFlag
	"00000000" + // _oBreak + alignment
	"01000000" + // _oSolidFlag
	"00000000" + // _oMissFlag
	"04000000" + // _oSelFlag + alignment
	"00000000" + // _oPreFlag
	"00000000" + // _oTrapFlag
	"00000000" + // _oDoorFlag
	"03000000" + // _olid
	"2a000000" + // _oRndSeed
	"00000000" + // _oVar1 zeroed (dynamic light list)
	"00000000000000000000000000000000" + // vars 2-5
	"00000000" + // _oVar6
	"00000000" + // bookMessage
	"00000000" // _oVar8

// roundtripLight is the populated light the roundtrip tests encode.
var roundtripLight = persist.LightRecord{
	X: 3, Y: 4, Radius: 15, IsInvalid: true, HasChanged: false,
	OldX: 1, OldY: 2, OldRadius: 7, OffsetX: 5, OffsetY: 6,
}

// TestLightRecordImage pins the SaveLighting record bytes for a standing
// light (loadsave.cpp:1795-1812).
func TestLightRecordImage(t *testing.T) {
	// given a standing light
	light := persist.LightRecord{
		X: 12, Y: 34, Radius: 9,
		IsInvalid: false, HasChanged: true,
		OldX: 11, OldY: 35, OldRadius: 8,
		OffsetX: -1, OffsetY: 2,
	}

	// when it is encoded
	got := light.Encode()

	// then the bytes match the pin image
	want, _ := hex.DecodeString(lightingWantHex)
	if !bytes.Equal(got, want) {
		t.Fatalf("light image = %x, want %x", got, want)
	}
}

// TestLightRecordVisionFlag pins the vision=true variant: _lid and _lflags
// both read 1 (loadsave.cpp:1810-1812).
func TestLightRecordVisionFlag(t *testing.T) {
	// given a vision light with all other fields zero
	light := persist.LightRecord{}

	// when it is encoded as vision
	got := light.EncodeVision()

	// then _lid and _lflags are 1 and everything else is zero
	want := make([]byte, 52)
	binary.LittleEndian.PutUint32(want[12:16], 1)
	binary.LittleEndian.PutUint32(want[48:52], 1)
	if !bytes.Equal(got, want) {
		t.Fatalf("vision image = %x, want %x", got, want)
	}
}

// TestLightRecordRoundTrip checks encode->decode->encode is stable for the
// standing variant.
func TestLightRecordRoundTrip(t *testing.T) {
	// given the standing image of a populated light
	image := roundtripLight.Encode()

	// when it is decoded and re-encoded
	reencoded := decodeReencodeLight(t, image)

	// then it is byte-identical
	if !bytes.Equal(reencoded, image) {
		t.Fatalf("re-encode changed the image: %x != %x", reencoded, image)
	}
}

// TestLightRecordVisionRoundTrip checks the same stability for the vision
// variant.
func TestLightRecordVisionRoundTrip(t *testing.T) {
	// given the vision image of an empty light
	image := persist.LightRecord{}.EncodeVision()

	// when it is decoded and re-encoded
	reencoded := decodeReencodeVision(t, image)

	// then it is byte-identical
	if !bytes.Equal(reencoded, image) {
		t.Fatalf("re-encode changed the image: %x != %x", reencoded, image)
	}
}

// TestPortalRecordImage pins the SavePortal record bytes
// (loadsave.cpp:1813-1822).
func TestPortalRecordImage(t *testing.T) {
	// given a closed portal
	portal := persist.PortalRecord{Open: false, X: 75, Y: 68, Level: 4, LType: 1, SetLevel: false}

	// when it is encoded
	got := portal.Encode()

	// then the bytes match the pin image
	want, _ := hex.DecodeString(portalWantHex)
	if !bytes.Equal(got, want) {
		t.Fatalf("portal image = %x, want %x", got, want)
	}
}

// TestPortalRecordSetLevelType pins the set-level branch: the raw ltype is
// written when setlvl is set (loadsave.cpp:1820).
func TestPortalRecordSetLevelType(t *testing.T) {
	// given an open set-level portal
	portal := persist.PortalRecord{Open: true, X: 1, Y: 2, Level: 5, LType: 4, SetLevel: true}

	// when it is encoded
	got := portal.Encode()

	// then open and ltype carry their raw values
	want, _ := hex.DecodeString("010000000100000002000000050000000400000001000000")
	if !bytes.Equal(got, want) {
		t.Fatalf("set-level portal image = %x, want %x", got, want)
	}
}

// TestPortalRecordRoundTrip checks encode->decode->encode stability.
func TestPortalRecordRoundTrip(t *testing.T) {
	// given a portal on both set-level branches
	for _, portal := range []persist.PortalRecord{
		{Open: true, X: 10, Y: 20, Level: 3, LType: 2, SetLevel: false},
		{Open: false, X: 11, Y: 21, Level: 9, LType: 4, SetLevel: true},
	} {
		// when the record is decoded from its own image
		image := portal.Encode()
		got, err := persist.DecodePortalRecord(image)
		if err != nil {
			t.Fatalf("DecodePortalRecord: %v", err)
		}

		// then the record survives and re-encodes identically
		if got != portal {
			t.Fatalf("decode = %+v, want %+v", got, portal)
		}

		if again := got.Encode(); !bytes.Equal(again, image) {
			t.Fatalf("re-encode changed the image: %x != %x", again, image)
		}
	}
}

// TestDecodePortalRecordShort asserts the short-image error.
func TestDecodePortalRecordShort(t *testing.T) {
	// when a truncated image is decoded
	_, err := persist.DecodePortalRecord(make([]byte, 8))

	// then the decode fails
	if err == nil {
		t.Fatal("DecodePortalRecord(truncated) = nil, want error")
	}
}

// TestObjectRecordImage pins the SaveObject record bytes with the
// dynamic-light var1 zeroing (loadsave.cpp:1749-1763).
func TestObjectRecordImage(t *testing.T) {
	// given a torch object whose var1 is the live light id 5
	object := persist.ObjectRecord{
		Type: 1157, X: 63, Y: 62, ApplyLighting: true,
		AnimDelay: 5, AnimLen: 10, AnimFrame: 1, AnimWidth: 96,
		SolidFlag: true, SelectionRegion: 4,
		OLid: 3, ORndSeed: 42,
		Var1: 5, IsDynamicLight: true,
	}

	// when it is encoded
	got := object.Encode()

	// then the bytes match the pin image, var1 zeroed
	want, err := hex.DecodeString(objectWantHex)
	if err != nil {
		t.Fatalf("fixture hex: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("object image = %x, want %x", got, want)
	}
}

// TestObjectRecordKeepsNonDynamicVar1 pins the default branch: var1 outside
// the dynamic-light object set is written unchanged (loadsave.cpp:1760-1762).
func TestObjectRecordKeepsNonDynamicVar1(t *testing.T) {
	// given a lever object with var1 = -7 and no dynamic-light flag
	object := persist.ObjectRecord{Type: 3, Var1: -7}

	// when it is encoded
	got := object.Encode()

	// then var1 is at its record offset, unchanged
	var1 := int32(uint32(got[88]) | uint32(got[89])<<8 | uint32(got[90])<<16 | uint32(got[91])<<24)
	if var1 != -7 {
		t.Fatalf("var1 = %d, want -7", var1)
	}
}

// TestObjectRecordRoundTrip checks encode->decode->encode stability.
func TestObjectRecordRoundTrip(t *testing.T) {
	// given a fully-populated object
	object := persist.ObjectRecord{
		Type: 42, X: 5, Y: 6, ApplyLighting: true, AnimFlag: true,
		AnimDelay: 2, AnimCnt: 3, AnimLen: 8, AnimFrame: 4, AnimWidth: 64,
		DelFlag: true, Break: 1, SolidFlag: true, MissFlag: true,
		SelectionRegion: 3, PreFlag: true, TrapFlag: true, DoorFlag: true,
		OLid: 9, ORndSeed: 77, Var1: 1, Var2: 2, Var3: 3, Var4: 4, Var5: 5,
		Var6: 6, BookMessage: 7, Var8: 8,
	}

	// when the record is decoded from its own image
	image := object.Encode()
	got, err := persist.DecodeObjectRecord(image)
	if err != nil {
		t.Fatalf("DecodeObjectRecord: %v", err)
	}

	// then every field survives
	if got != object {
		t.Fatalf("decode = %+v, want %+v", got, object)
	}

	if again := got.Encode(); !bytes.Equal(again, image) {
		t.Fatalf("re-encode changed the image: %x != %x", again, image)
	}
}

// TestDecodeObjectRecordShort asserts the short-image error.
func TestDecodeObjectRecordShort(t *testing.T) {
	// when a truncated image is decoded
	_, err := persist.DecodeObjectRecord(make([]byte, 40))

	// then the decode fails
	if err == nil {
		t.Fatal("DecodeObjectRecord(truncated) = nil, want error")
	}
}

// TestObjectRecordSize asserts the 120-byte record stride.
func TestObjectRecordSize(t *testing.T) {
	// given an object record
	object := persist.ObjectRecord{}

	// when it is encoded
	got := object.Encode()

	// then it is ObjectSaveSize bytes (120, memory_map object.txt)
	if len(got) != persist.ObjectSaveSize {
		t.Fatalf("object image = %d bytes, want %d", len(got), persist.ObjectSaveSize)
	}
}

// TestDecodeLightRecordShort asserts the short-image error.
func TestDecodeLightRecordShort(t *testing.T) {
	// when a truncated image is decoded
	_, _, err := persist.DecodeLightRecord(make([]byte, 12))

	// then the decode fails with ErrShortRecord
	if err == nil {
		t.Fatal("DecodeLightRecord(truncated) = nil, want error")
	}
}

// decodeReencodeLight decodes a standing image and returns the re-encode.
func decodeReencodeLight(t *testing.T, image []byte) []byte {
	t.Helper()

	light, _, err := persist.DecodeLightRecord(image)
	if err != nil {
		t.Fatalf("DecodeLightRecord: %v", err)
	}

	return light.Encode()
}

// decodeReencodeVision decodes a vision image and returns the re-encode.
func decodeReencodeVision(t *testing.T, image []byte) []byte {
	t.Helper()

	light, vision, err := persist.DecodeLightRecord(image)
	if err != nil {
		t.Fatalf("DecodeLightRecord: %v", err)
	}

	if !vision {
		t.Fatal("decode lost the vision variant")
	}

	return light.EncodeVision()
}
