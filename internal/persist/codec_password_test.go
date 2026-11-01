package persist_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestHeroEntryPasswordRoundTrip pins the save-mode passwords: upstream encodes
// and decodes the hero entry with pfile_get_password(), which returns a
// different password per mode (Source/pfile.cpp:638-643), so an entry written
// in one mode has to be unreadable in the others.
func TestHeroEntryPasswordRoundTrip(t *testing.T) {
	t.Parallel()

	pack := goldenPlayerPack().Encode()

	for _, password := range heroPasswords() {
		t.Run(password, func(t *testing.T) {
			t.Parallel()

			// when an entry is written with one mode password
			entry := encodeHeroEntry(t, pack, password)

			// then it reads back and a foreign password is rejected
			requireEntryRoundTrip(t, entry, pack, password)
		})
	}
}

// encodeHeroEntry writes a hero entry with one mode's password.
func encodeHeroEntry(t *testing.T, pack []byte, password string) []byte {
	t.Helper()

	entry, err := persist.EncodeHeroEntry(pack, password)
	if err != nil {
		t.Fatalf("EncodeHeroEntry: %v", err)
	}

	return entry
}

// requireEntryRoundTrip checks one password round-trips and that a foreign
// mode's password is rejected.
func requireEntryRoundTrip(t *testing.T, entry, pack []byte, password string) {
	t.Helper()

	decoded, err := persist.DecodeHeroEntry(entry, password)
	if err != nil {
		t.Fatalf("DecodeHeroEntry with %q: %v", password, err)
	}

	requirePackBytes(t, decoded, pack)
	requireWrongPasswordRejected(t, entry, otherPassword(password))
}

// heroPasswords are the four pfile_get_password values (Source/pfile.cpp:60-63).
func heroPasswords() []string {
	return []string{
		persist.PasswordSingle,
		persist.PasswordMulti,
		persist.PasswordSpawnSingle,
		persist.PasswordSpawnMulti,
	}
}

// otherPassword returns a password of a different mode.
func otherPassword(password string) string {
	for _, candidate := range heroPasswords() {
		if candidate != password {
			return candidate
		}
	}

	return password
}

// requirePackBytes fails unless the decoded payload is the original pack.
func requirePackBytes(t *testing.T, got, want []byte) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("decoded len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("decoded byte %d = %#x, want %#x", i, got[i], want[i])
		}
	}
}

// requireWrongPasswordRejected fails when the codec accepts a foreign password.
func requireWrongPasswordRejected(t *testing.T, entry []byte, wrong string) {
	t.Helper()

	// the codec decrypts in place, so hand it a copy
	if _, err := persist.DecodeHeroEntry(append([]byte(nil), entry...), wrong); err == nil {
		t.Fatalf("DecodeHeroEntry accepted password %q", wrong)
	}
}
