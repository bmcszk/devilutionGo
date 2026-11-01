package persist_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestHeroEntryLen(t *testing.T) {
	// when the codec length of a hero record is computed (codec_test.cpp:7-16)
	cases := []struct{ payload, want int }{
		{50, 72},
		{128, 136},
		{persist.PlayerPackSize, 1288},
	}
	// then it pads to the 32-byte block and adds the 8-byte signature
	for _, tc := range cases {
		if got := persist.HeroEntryLen(tc.payload); got != tc.want {
			t.Fatalf("HeroEntryLen(%d) = %d, want %d", tc.payload, got, tc.want)
		}
	}
}

func TestHeroPassword(t *testing.T) {
	// when each mode picks its password (pfile.cpp:638-643)
	cases := []struct {
		spawn, multi bool
		want         string
	}{
		{false, false, persist.PasswordSingle},
		{false, true, persist.PasswordMulti},
		{true, false, persist.PasswordSpawnSingle},
		{true, true, persist.PasswordSpawnMulti},
	}
	// then the pfile constants match
	for _, tc := range cases {
		mode := persist.SaveNamespace{Spawn: tc.spawn, Multiplayer: tc.multi}
		got := persist.HeroPassword(mode)

		if got != tc.want {
			t.Fatalf("HeroPassword(%v,%v) = %q, want %q", tc.spawn, tc.multi, got, tc.want)
		}
	}
}

func TestHeroEntryRoundTrip(t *testing.T) {
	// given a full hero record
	payload := goldenPlayerPack().Encode()
	// when it is encoded and decoded through the codec
	encoded, err := persist.EncodeHeroEntry(payload, persist.PasswordSingle)
	if err != nil {
		t.Fatalf("EncodeHeroEntry: %v", err)
	}

	// then the ciphertext differs from the record (checked before the decode,
	// which decrypts in place) and the payload round-trips
	if bytes.Equal(encoded[:len(payload)], payload) {
		t.Fatal("encoded entry is not encrypted")
	}

	decoded, err := persist.DecodeHeroEntry(encoded, persist.PasswordSingle)
	if err != nil {
		t.Fatalf("DecodeHeroEntry: %v", err)
	}

	if !bytes.Equal(decoded, payload) {
		t.Fatal("decoded payload differs from the record")
	}

	if len(encoded) != persist.HeroEntryLen(len(payload)) {
		t.Fatalf("encoded len = %d, want %d", len(encoded), persist.HeroEntryLen(len(payload)))
	}
}

func TestDecodeHeroEntryWrongPassword(t *testing.T) {
	// given an entry encoded with the single-player password
	payload := goldenPlayerPack().Encode()
	encoded, err := persist.EncodeHeroEntry(payload, persist.PasswordSingle)
	if err != nil {
		t.Fatalf("EncodeHeroEntry: %v", err)
	}
	// when it is decoded with the multiplayer password
	_, err = persist.DecodeHeroEntry(encoded, persist.PasswordMulti)
	// then the codec rejects it
	if !errors.Is(err, persist.ErrCodecRejected) {
		t.Fatalf("err = %v, want ErrCodecRejected", err)
	}
}

func TestEncodeHeroEntryEmptyPassword(t *testing.T) {
	// when the password is empty
	_, err := persist.EncodeHeroEntry(make([]byte, 10), "")
	// then the call is refused before the codec runs
	if !errors.Is(err, persist.ErrEmptyPassword) {
		t.Fatalf("err = %v, want ErrEmptyPassword", err)
	}
}
