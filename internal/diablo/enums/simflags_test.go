package enums_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
)

// UseMultiplayerQuests is fullQuests == 0, and fullQuests is
// (!gbIsMultiplayer || multiplayerFullQuests) ? 1 : 0
// (questdat.cpp:59-62, multi.cpp:614).
func TestUseMultiplayerQuests(t *testing.T) {
	tests := []struct {
		name    string
		flags   enums.SimFlags
		want    bool
		because string
	}{
		{
			name:    "singlePlayer",
			flags:   enums.SimFlags{},
			want:    false,
			because: "single player always uses the retail quest table",
		},
		{
			name:    "multiplayer",
			flags:   enums.SimFlags{Multiplayer: true},
			want:    true,
			because: "multiplayer without the full-quests option sets fullQuests 0",
		},
		{
			name:    "multiplayerFullQuests",
			flags:   enums.SimFlags{Multiplayer: true, FullQuests: true},
			want:    false,
			because: "multiplayerFullQuests raises fullQuests back to 1",
		},
		{
			name:    "fullQuestsWithoutMultiplayer",
			flags:   enums.SimFlags{FullQuests: true},
			want:    false,
			because: "the option is only read in multiplayer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := tt.flags.UseMultiplayerQuests()
			// then
			if got != tt.want {
				t.Fatalf("UseMultiplayerQuests() = %v, want %v: %s", got, tt.want, tt.because)
			}
		})
	}
}

// CheckQuests returns before touching any quest when gbIsSpawn is set
// (quests.cpp:135-136).
func TestQuestsEnabled(t *testing.T) {
	tests := []struct {
		name  string
		flags enums.SimFlags
		want  bool
	}{
		{name: "retail", flags: enums.SimFlags{}, want: true},
		{name: "hellfire", flags: enums.SimFlags{Hellfire: true}, want: true},
		{name: "multiplayer", flags: enums.SimFlags{Multiplayer: true}, want: true},
		{name: "spawn", flags: enums.SimFlags{Spawn: true}, want: false},
		{name: "spawnHellfire", flags: enums.SimFlags{Spawn: true, Hellfire: true}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := tt.flags.QuestsEnabled()
			// then
			if got != tt.want {
				t.Fatalf("QuestsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
