// Package appfat ports Source/appfat.cpp: the terminating application-error
// surface (app_fatal, ErrDlg, InsertCDDlg, DirErrorDlg) as typed errors.
//
// Upstream terminates inside the dialog (diablo_quit(1)); the Go port returns
// a *FatalError so the cmd boundary owns the exit, matching the layer law that only
// entry points may end the process. FreeDlg's multiplayer cleanup
// (appfat.cpp:44-58) has no equivalent until the net wave.
package appfat

import (
	"fmt"
	"io"
)

// FatalError is a terminating application error, ready for a platform presenter.
// Title and Body map to the UiErrorOkDialog arguments (appfat.cpp:62-67).
type FatalError struct {
	Title string
	Body  string

	cause error
}

// New ports the DisplayFatalErrorAndExit payload (appfat.cpp:62-67).
func New(title, body string) *FatalError {
	return &FatalError{Title: title, Body: body}
}

// Error renders the dialog caption and text for error chains.
func (f *FatalError) Error() string {
	return f.Title + ": " + f.Body
}

// Unwrap exposes the underlying cause, if the fatal came from ErrDlg.
func (f *FatalError) Unwrap() error {
	return f.cause
}

// WriteTo emits the caption then the body, the headless presentation path:
// UiOkDialog always LogErrors "{caption}\n{text}" and skips the message box
// when HeadlessMode is set (dialogs.cpp:170-181).
func (f *FatalError) WriteTo(w io.Writer) (int64, error) {
	message := f.Title + "\n" + f.Body + "\n"

	n, err := io.WriteString(w, message)
	if err != nil {
		return int64(n), fmt.Errorf("appfat: write %q: %w", f.Title, err)
	}

	return int64(n), nil
}

// AppFatal ports app_fatal, which always uses the "Error" caption
// (appfat.cpp:69-72).
func AppFatal(body string) *FatalError {
	return New("Error", body)
}

// AssertionFailed ports the _DEBUG assert_fail body (appfat.cpp:74-79).
func AssertionFailed(file string, line int, expr string) *FatalError {
	return AppFatal(fmt.Sprintf("assertion failed (%s:%d)\n%s", file, line, expr))
}

// InsertCDDlg ports the missing-archive error (appfat.cpp:87-93).
func InsertCDDlg(archiveName string) *FatalError {
	return New("Data File Error", fmt.Sprintf(
		"Unable to open main data archive (%s).\n\nMake sure that it is in the game folder.", archiveName))
}

// DirErrorDlg ports the read-only-directory error (appfat.cpp:94-100).
func DirErrorDlg(dir string) *FatalError {
	return New("Read-Only Directory Error", fmt.Sprintf("Unable to write to location:\n%s", dir))
}

// ErrDlg ports ErrDlg, appending the "occurred at" location line the C++
// format string adds (appfat.cpp:80-86). The cause is wrapped, not inlined, so
// callers keep errors.Is chains.
func ErrDlg(title string, cause error, file string, line int) *FatalError {
	fatal := New(title, fmt.Sprintf("%s\n\nThe error occurred at: %s line %d", cause, file, line))
	fatal.cause = cause

	return fatal
}
