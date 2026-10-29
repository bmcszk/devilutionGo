package appdiag_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/appdiag"
	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
)

// The bodies are the upstream app_fatal strings verbatim; a drift here changes
// the bytes a player sees, so each one is pinned.
func TestMessages_MatchUpstreamAppFatalText(t *testing.T) {
	tests := []struct {
		name      string
		fatal     *appfat.FatalError
		wantTitle string
		wantBody  string
	}{
		{
			"invalid encode parameters",
			appdiag.InvalidEncodeParameters(),
			"Error",
			"Invalid encode parameters",
		},
		{
			"wrong missile id",
			appdiag.WrongMissileID(17),
			"Error",
			"wrong missile ID 17",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			fatal := test.fatal

			// when
			title, body := fatal.Title, fatal.Body

			// then
			assertMessage(t, title, body, test.wantTitle, test.wantBody)
		})
	}
}

func TestMessages_AreFatalErrors(t *testing.T) {
	// given
	want := appdiag.InvalidEncodeParameters()

	// when
	var got *appfat.FatalError

	// then
	if !errors.As(error(want), &got) {
		t.Fatalf("codec message is not a *appfat.FatalError: %T", want)
	}
}

// assertMessage pins one ported app_fatal message pair.
func assertMessage(t *testing.T, title, body, wantTitle, wantBody string) {
	t.Helper()

	if title != wantTitle {
		t.Fatalf("title = %q, want %q (app_fatal always uses \"Error\")", title, wantTitle)
	}

	if body != wantBody {
		t.Fatalf("body = %q, want %q", body, wantBody)
	}
}
