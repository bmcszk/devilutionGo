package appfat_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
)

// errSink is a writer that always fails, standing in for a closed stderr.
type errSink struct{ err error }

func (s errSink) Write([]byte) (int, error) { return 0, s.err }

// Ports test/appfat_test.cpp:9-10 — app_fatal("test") exits 1 with body "test".
func TestAppFatal_Message_UsesErrorTitleAndBody(t *testing.T) {
	// given
	fatal := appfat.AppFatal("test")

	// when
	_ = fatal

	// then
	if fatal.Title != "Error" {
		t.Fatalf("app_fatal title = %q, want %q (appfat.cpp:71 _(\"Error\"))", fatal.Title, "Error")
	}

	if fatal.Body != "test" {
		t.Fatalf("app_fatal body = %q, want %q (appfat.cpp:69-72)", fatal.Body, "test")
	}
}

// Ports test/appfat_test.cpp:14-17 — ErrDlg appends the "occurred at" line.
func TestErrDlg_Message_AppendsFileAndLine(t *testing.T) {
	// given
	want := "Unknown error\n\nThe error occurred at: appfat.cpp line 7"

	// when
	fatal := appfat.ErrDlg("Title", errors.New("Unknown error"), "appfat.cpp", 7)

	// then
	if got := fatal.Body; got != want {
		t.Fatalf("ErrDlg body = %q, want %q (appfat.cpp:80-86)", got, want)
	}

	if fatal.Title != "Title" {
		t.Fatalf("ErrDlg title = %q, want %q", fatal.Title, "Title")
	}
}

func TestErrDlg_Message_WrapsCause(t *testing.T) {
	// given
	cause := errors.New("disk on fire")

	// when
	fatal := appfat.ErrDlg("Title", cause, "save.go", 42)

	// then
	if !errors.Is(fatal, cause) {
		t.Fatalf("ErrDlg(%v) lost its cause; want errors.Is(fatal, cause) (golang-error-handling: wrap with %%w)", fatal)
	}
}

// Ports test/appfat_test.cpp:19-22 — InsertCDDlg names the missing archive.
func TestInsertCDDlg_Message_NamesArchive(t *testing.T) {
	// given
	want := "Unable to open main data archive (diabdat.mpq).\n\nMake sure that it is in the game folder."

	// when
	fatal := appfat.InsertCDDlg("diabdat.mpq")

	// then
	if got := fatal.Body; got != want {
		t.Fatalf("InsertCDDlg body = %q, want %q (appfat.cpp:87-93)", got, want)
	}

	if fatal.Title != "Data File Error" {
		t.Fatalf("InsertCDDlg title = %q, want %q", fatal.Title, "Data File Error")
	}
}

// Ports test/appfat_test.cpp:24-27 — DirErrorDlg reports the unwritable path.
func TestDirErrorDlg_Message_ReportsPath(t *testing.T) {
	// given
	want := "Unable to write to location:\n/"

	// when
	fatal := appfat.DirErrorDlg("/")

	// then
	if got := fatal.Body; got != want {
		t.Fatalf("DirErrorDlg body = %q, want %q (appfat.cpp:94-100)", got, want)
	}

	if fatal.Title != "Read-Only Directory Error" {
		t.Fatalf("DirErrorDlg title = %q, want %q", fatal.Title, "Read-Only Directory Error")
	}
}

// Ports appfat.cpp:74-79 — the _DEBUG assert_fail body format.
func TestAssertionFailed_Message_MatchesDebugAssert(t *testing.T) {
	// given
	want := "assertion failed (palette.go:12)\nlen(map) > 0"

	// when
	fatal := appfat.AssertionFailed("palette.go", 12, "len(map) > 0")

	// then
	if got := fatal.Body; got != want {
		t.Fatalf("assert body = %q, want %q (appfat.cpp:77)", got, want)
	}
}

// Headless has no dialog: the message goes to stderr (dialogs.cpp:170-181
// LogError + HeadlessMode skip).
func TestWriteTo_Stderr_EmitsTitleThenBody(t *testing.T) {
	tests := []struct {
		name  string
		fatal *appfat.FatalError
		want  string
	}{
		{"app fatal", appfat.AppFatal("test"), "Error\ntest\n"},
		{
			"err dialog",
			appfat.ErrDlg("Title", errors.New("Unknown error"), "appfat.cpp", 7),
			"Title\nUnknown error\n\nThe error occurred at: appfat.cpp line 7\n",
		},
		{"insert cd", appfat.InsertCDDlg("diabdat.mpq"), "Data File Error\n" +
			"Unable to open main data archive (diabdat.mpq).\n\n" +
			"Make sure that it is in the game folder.\n"},
		{"dir error", appfat.DirErrorDlg("/"), "Read-Only Directory Error\nUnable to write to location:\n/\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			fatal := test.fatal

			// when
			got, n, err := writtenFatal(fatal)

			// then
			assertWritten(t, got, n, err, test.want)
		})
	}
}

func TestWriteTo_SinkFailure_ReturnsWrappedError(t *testing.T) {
	// given
	sink := errSink{err: errors.New("broken pipe")}

	// when
	_, err := appfat.New("Title", "Body").WriteTo(sink)

	// then
	if err == nil {
		t.Fatal("WriteTo() error = nil, want the sink failure (golang-error-handling: propagate, do not swallow)")
	}

	if !errors.Is(err, sink.err) {
		t.Fatalf("WriteTo() error = %v, want it to wrap the sink error", err)
	}
}

// writtenFatal presents a fatal error into a fresh buffer.
func writtenFatal(fatal *appfat.FatalError) (string, int64, error) {
	var buf bytes.Buffer

	n, err := fatal.WriteTo(&buf)

	return buf.String(), int64(n), err
}

// assertWritten checks the exact headless presentation of a fatal error.
func assertWritten(t *testing.T, got string, n int64, err error, want string) {
	t.Helper()

	if err != nil {
		t.Fatalf("WriteTo() error = %v, want nil", err)
	}

	if got != want {
		t.Fatalf("WriteTo() wrote %q, want %q", got, want)
	}

	if int(n) != len(want) {
		t.Fatalf("WriteTo() reported %d bytes, want %d", n, len(want))
	}
}
